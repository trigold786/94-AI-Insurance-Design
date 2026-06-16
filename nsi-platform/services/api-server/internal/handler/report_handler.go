package handler

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type PolicySearcher interface {
	QueryByRegionAndStatus(ctx context.Context, regionCode, status string) ([]models.PolicyClaim, error)
}

func PlanReportHandler(repo PlanRepository, policyRepo PolicySearcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)

		planID := r.URL.Query().Get("plan_id")
		if planID == "" {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "plan_id required"})
			return
		}

		plan, err := repo.GetByID(r.Context(), planID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				respondJSON(w, http.StatusNotFound, map[string]interface{}{"code": "NOT_FOUND", "message": "plan not found"})
				return
			}
			respondError(w, err)
			return
		}

		if plan.UserID != userID {
			respondJSON(w, http.StatusNotFound, map[string]interface{}{"code": "NOT_FOUND", "message": "plan not found"})
			return
		}

		var policies []models.PolicyClaim
		if plan.RecommendedSchemes != nil && len(plan.RecommendedSchemes) > 0 && policyRepo != nil {
			if p, err := policyRepo.QueryByRegionAndStatus(r.Context(), "", "verified"); err == nil {
				policies = p
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		renderPlanReport(w, plan, policies)
	})
}

type reportData struct {
	Plan        *models.PlanSnapshot
	Policies    []models.PolicyClaim
	GeneratedAt string
}

func renderPlanReport(w http.ResponseWriter, plan *models.PlanSnapshot, policies []models.PolicyClaim) {
	data := reportData{
		Plan:        plan,
		Policies:    policies,
		GeneratedAt: time.Now().Format("2006年01月02日 15:04"),
	}

	tmpl := template.Must(template.New("report").Parse(reportHTML))
	tmpl.Execute(w, data)
}

func PlanReportPDFHandler(repo PlanRepository, policyRepo PolicySearcher, profileRepo ProfileLookuper) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)

		planID := r.URL.Query().Get("plan_id")
		if planID == "" {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "plan_id required"})
			return
		}

		plan, err := repo.GetByID(r.Context(), planID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				respondJSON(w, http.StatusNotFound, map[string]interface{}{"code": "NOT_FOUND", "message": "plan not found"})
				return
			}
			respondError(w, err)
			return
		}

		if plan.UserID != userID {
			respondJSON(w, http.StatusNotFound, map[string]interface{}{"code": "NOT_FOUND", "message": "plan not found"})
			return
		}

		var policies []models.PolicyClaim
		if plan.RecommendedSchemes != nil && len(plan.RecommendedSchemes) > 0 && policyRepo != nil {
			if p, err := policyRepo.QueryByRegionAndStatus(r.Context(), "", "verified"); err == nil {
				policies = p
			}
		}

		var profile *models.UserProfile
		if profileRepo != nil {
			if p, err := profileRepo.GetByUserID(r.Context(), userID); err == nil {
				profile = p
			}
		}

		cityName := ""
		if profile != nil {
			code := profile.CurrentResidenceCode
			if code == "" {
				code = profile.HouseholdRegionCode
			}
			if ci := GetCityInfo(code); ci != nil {
				cityName = ci.Name
			}
		}

		var htmlBuf bytes.Buffer
		renderPDFReport(&htmlBuf, plan, policies, profile, cityName)
		htmlFile := "/tmp/nsi-report-" + planID + ".html"
		pdfFile := "/tmp/nsi-report-" + planID + ".pdf"
		os.WriteFile(htmlFile, htmlBuf.Bytes(), 0644)
		chromeBin := os.Getenv("CHROME_BIN")
		if chromeBin == "" {
			chromeBin = "chromium-browser"
		}
		cmd := exec.Command(chromeBin, "--headless", "--no-sandbox", "--disable-gpu",
			"--print-to-pdf="+pdfFile, htmlFile)
		cmd.Stdout = nil
		cmd.Stderr = nil
		pdfErr := cmd.Run()
		if pdfErr == nil {
			pdfData, readErr := os.ReadFile(pdfFile)
			if readErr == nil && len(pdfData) > 0 {
				w.Header().Set("Content-Type", "application/pdf")
				w.Header().Set("Content-Disposition", `attachment; filename="social-insurance-report.pdf"`)
				w.Write(pdfData)
				os.Remove(htmlFile)
				os.Remove(pdfFile)
				return
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="social-insurance-report.html"`)
		w.Write(htmlBuf.Bytes())
	})
}

type pdfReportData struct {
	Plan        *models.PlanSnapshot
	Policies    []models.PolicyClaim
	Profile     *models.UserProfile
	CityName    string
	GeneratedAt string
}

func renderPDFReport(w io.Writer, plan *models.PlanSnapshot, policies []models.PolicyClaim, profile *models.UserProfile, cityName string) {
	data := pdfReportData{
		Plan:        plan,
		Policies:    policies,
		Profile:     profile,
		CityName:    cityName,
		GeneratedAt: time.Now().Format("2006年01月02日 15:04"),
	}

	tmpl := template.Must(template.New("pdfreport").Parse(reportPDFHTML))
	tmpl.Execute(w, data)
}

const reportHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>社保规划报告 - AI社保智筹</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:"Microsoft YaHei","PingFang SC",sans-serif;color:#333;background:#fff;padding:20px;font-size:14px;line-height:1.6}
@media print{body{padding:0}}
h1{font-size:22px;color:#1A56DB;margin-bottom:4px}
h2{font-size:16px;color:#1A56DB;border-bottom:2px solid #1A56DB;padding-bottom:6px;margin:24px 0 12px}
h3{font-size:14px;color:#374151;margin:12px 0 6px}
.header{text-align:center;padding:20px 0 16px;border-bottom:2px solid #E5E7EB;margin-bottom:20px}
.header .meta{color:#6B7280;font-size:13px;margin-top:6px}
table{width:100%;border-collapse:collapse;margin:8px 0 16px;font-size:13px}
th{background:#F3F4F6;text-align:left;padding:8px 10px;border-bottom:2px solid #E5E7EB;color:#6B7280;font-weight:600;font-size:12px}
td{padding:7px 10px;border-bottom:1px solid #F3F4F6}
tr:hover{background:#F9FAFB}
.highlight{background:#EFF6FF;font-weight:600}
.badge{display:inline-block;padding:1px 6px;border-radius:4px;font-size:11px;font-weight:500}
.bg-green{background:#D1FAE5;color:#059669}
.bg-yellow{background:#FEF3C7;color:#D97706}
.steps{padding-left:20px}
.steps li{margin:6px 0;padding:4px 0}
.footer{text-align:center;color:#9CA3AF;font-size:12px;border-top:1px solid #E5E7EB;padding:12px 0;margin-top:24px}
.scheme-card{background:#F9FAFB;border-radius:8px;padding:12px;margin:8px 0}
.scheme-card h4{color:#1A56DB;font-size:14px;margin-bottom:4px}
</style>
</head>
<body>

<div class="header">
<h1>社保规划报告</h1>
<div class="meta">生成时间: {{.GeneratedAt}} | 方案ID: {{.Plan.PlanID}}</div>
</div>

<h2>方案概览</h2>
<table>
<tr><th>方案</th><th>缴费基数</th><th>月缴金额</th><th>年补贴</th><th>预计月养老金</th></tr>
{{range .Plan.RecommendedSchemes}}
<tr>
<td>{{.Name}}</td>
<td>{{.BaseSalary}} 元</td>
<td>{{printf "%.2f" .MonthlyCost}} 元</td>
<td>{{printf "%.2f" .AnnualSubsidy}} 元/年</td>
<td>{{printf "%.2f" .ProjectedPension}} 元</td>
</tr>
{{end}}
</table>
<div style="background:#EFF6FF;border-radius:8px;padding:12px;margin:12px 0">
<div style="font-size:13px;color:#374151">终身总投入: <strong>{{printf "%.2f" .Plan.TotalCost}} 元</strong></div>
<div style="font-size:13px;color:#374151">预计政府补贴: <strong>{{printf "%.2f" .Plan.TotalSubsidy}} 元</strong></div>
</div>

{{if .Plan.RecommendedSchemes}}
{{$first := index .Plan.RecommendedSchemes 0}}
{{if $first.Cashflow}}
<h2>现金流预测</h2>
<table>
<tr><th>年份</th><th>年缴费</th><th>年补贴</th><th>账户余额</th></tr>
{{range $first.Cashflow}}
<tr>
<td>{{.Year}} 年</td>
<td>{{printf "%.2f" .Payment}} 元</td>
<td>{{printf "%.2f" .Subsidy}} 元</td>
<td>{{printf "%.2f" .Balance}} 元</td>
</tr>
{{end}}
</table>
{{end}}
{{end}}

<h2>推荐方案详情</h2>
{{range .Plan.RecommendedSchemes}}
<div class="scheme-card">
<h4>{{.Name}}</h4>
<div>缴费基数: {{.BaseSalary}} 元/月</div>
<div>月缴金额: {{printf "%.2f" .MonthlyCost}} 元</div>
<div>预计月养老金: {{printf "%.2f" .ProjectedPension}} 元</div>
</div>
{{end}}

<h2>权益说明</h2>
<table>
<tr><th>险种</th><th>说明</th></tr>
<tr><td>养老保险</td><td>累计缴费满15年，达到法定退休年龄后按月领取养老金。缴费基数越高、年限越长，领取金额越高。</td></tr>
<tr><td>医疗保险</td><td>享受住院、门诊等医疗费用报销。灵活就业人员可参加职工医保或居民医保。</td></tr>
<tr><td>失业保险</td><td>缴费满1年可申领失业金，最长24个月。灵活就业人员需符合当地政策。</td></tr>
<tr><td>工伤保险</td><td>工作期间发生意外伤害可享受工伤待遇。灵活就业人员需单独参保。</td></tr>
</table>

{{if .Policies}}
<h2>适用政策清单</h2>
<table>
<tr><th>政策ID</th><th>类型</th><th>补贴计算</th><th>来源</th></tr>
{{range .Policies}}
<tr>
<td>{{.PolicyID}}</td>
<td>{{.PolicyType}}</td>
<td>{{.SubsidyCalcMethod}}</td>
<td>{{.SourceName}}</td>
</tr>
{{end}}
</table>
{{end}}

<h2>行动步骤</h2>
<ol class="steps">
<li>确认个人身份信息：准备好身份证、户口本、居住证等材料</li>
<li>选择参保方案：根据经济能力选择合适的缴费基数档次</li>
<li>办理参保登记：通过当地人社局官网或社保服务大厅办理灵活就业人员参保登记</li>
<li>按时缴费：每月按时足额缴纳社保费用，避免断缴影响待遇</li>
<li>定期查看权益：通过本平台跟踪缴费记录和权益状态</li>
</ol>

<div class="footer">
<p>本报告由 AI社保智筹 生成，仅供参考。具体政策以当地人社局官方发布为准。</p>
<p>AI社保智筹 - 智能社保规划平台</p>
</div>

</body>
</html>`

const reportPDFHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<title>社保筹划报告 - AI社保智筹</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:"Microsoft YaHei","PingFang SC",sans-serif;color:#333;background:#fff;padding:30px;font-size:13px;line-height:1.6}
@media print{body{padding:0}@page{size:A4;margin:15mm}}
h1{font-size:20px;color:#1A56DB;margin-bottom:4px;text-align:center}
h2{font-size:15px;color:#1A56DB;border-bottom:2px solid #1A56DB;padding-bottom:5px;margin:20px 0 10px;page-break-after:avoid}
h3{font-size:13px;color:#374151;margin:10px 0 5px}
.header{text-align:center;padding:16px 0 12px;border-bottom:2px solid #E5E7EB;margin-bottom:18px}
.header .meta{color:#6B7280;font-size:12px;margin-top:4px}
table{width:100%;border-collapse:collapse;margin:6px 0 14px;font-size:12px}
th{background:#F3F4F6;text-align:left;padding:6px 8px;border-bottom:2px solid #E5E7EB;color:#6B7280;font-weight:600}
td{padding:5px 8px;border-bottom:1px solid #F3F4F6}
tr:nth-child(even){background:#FAFBFC}
.highlight{background:#EFF6FF;font-weight:600}
.info-grid{display:grid;grid-template-columns:1fr 1fr;gap:6px;margin:8px 0 14px;font-size:12px}
.info-item{padding:6px 10px;background:#F9FAFB;border-radius:4px}
.info-label{color:#6B7280;font-size:11px}
.info-value{font-weight:600;color:#1F2937}
.scheme-card{background:#F9FAFB;border-radius:6px;padding:10px;margin:6px 0;page-break-inside:avoid}
.scheme-card h4{color:#1A56DB;font-size:13px;margin-bottom:3px}
.summary-box{background:#EFF6FF;border-radius:6px;padding:10px;margin:10px 0;font-size:12px}
.summary-box strong{color:#1A56DB}
.footer{text-align:center;color:#9CA3AF;font-size:11px;border-top:1px solid #E5E7EB;padding:10px 0;margin-top:20px}
</style>
</head>
<body>

<div class="header">
<h1>社保筹划报告</h1>
<div class="meta">生成时间: {{.GeneratedAt}} | 方案ID: {{.Plan.PlanID}}</div>
</div>

{{if .Profile}}
<h2>个人档案</h2>
<div class="info-grid">
<div class="info-item"><div class="info-label">年龄</div><div class="info-value">{{.Profile.Age}} 岁</div></div>
<div class="info-item"><div class="info-label">性别</div><div class="info-value">{{if eq .Profile.Gender "male"}}男{{else}}女{{end}}</div></div>
{{if .CityName}}<div class="info-item"><div class="info-label">所在城市</div><div class="info-value">{{.CityName}}</div></div>{{end}}
<div class="info-item"><div class="info-label">就业状态</div><div class="info-value">{{.Profile.EmploymentStatus}}</div></div>
</div>
{{end}}

<h2>方案概览</h2>
<table>
<tr><th>方案</th><th>缴费基数</th><th>月缴金额</th><th>年补贴</th><th>预计月养老金</th></tr>
{{range .Plan.RecommendedSchemes}}
<tr>
<td>{{.Name}}</td>
<td>{{.BaseSalary}} 元</td>
<td>{{printf "%.2f" .MonthlyCost}} 元</td>
<td>{{printf "%.2f" .AnnualSubsidy}} 元/年</td>
<td>{{printf "%.2f" .ProjectedPension}} 元</td>
</tr>
{{end}}
</table>

<div class="summary-box">
<div>终身总投入: <strong>{{printf "%.2f" .Plan.TotalCost}} 元</strong></div>
<div>预计政府补贴: <strong>{{printf "%.2f" .Plan.TotalSubsidy}} 元</strong></div>
</div>

{{if .Plan.RecommendedSchemes}}
{{$first := index .Plan.RecommendedSchemes 0}}
{{if $first.Cashflow}}
<h2>现金流预测</h2>
<table>
<tr><th>年份</th><th>年缴费</th><th>年补贴</th><th>账户余额</th></tr>
{{range $first.Cashflow}}
<tr>
<td>{{.Year}} 年</td>
<td>{{printf "%.2f" .Payment}} 元</td>
<td>{{printf "%.2f" .Subsidy}} 元</td>
<td>{{printf "%.2f" .Balance}} 元</td>
</tr>
{{end}}
</table>
{{end}}
{{end}}

<h2>推荐方案详情</h2>
{{range .Plan.RecommendedSchemes}}
<div class="scheme-card">
<h4>{{.Name}}</h4>
<div>缴费基数: {{.BaseSalary}} 元/月</div>
<div>月缴金额: {{printf "%.2f" .MonthlyCost}} 元</div>
<div>预计月养老金: {{printf "%.2f" .ProjectedPension}} 元</div>
</div>
{{end}}

<h2>权益说明</h2>
<table>
<tr><th>险种</th><th>说明</th></tr>
<tr><td>养老保险</td><td>累计缴费满15年，达到法定退休年龄后按月领取养老金。</td></tr>
<tr><td>医疗保险</td><td>享受住院、门诊等医疗费用报销。</td></tr>
<tr><td>失业保险</td><td>缴费满1年可申领失业金，最长24个月。</td></tr>
<tr><td>工伤保险</td><td>工作期间发生意外伤害可享受工伤待遇。</td></tr>
</table>

{{if .Policies}}
<h2>适用政策清单</h2>
<table>
<tr><th>政策ID</th><th>类型</th><th>补贴计算</th><th>来源</th></tr>
{{range .Policies}}
<tr>
<td>{{.PolicyID}}</td>
<td>{{.PolicyType}}</td>
<td>{{.SubsidyCalcMethod}}</td>
<td>{{.SourceName}}</td>
</tr>
{{end}}
</table>
{{end}}

<h2>行动步骤</h2>
<ol style="padding-left:18px;margin:6px 0">
<li>确认个人身份信息：准备好身份证、户口本、居住证等材料</li>
<li>选择参保方案：根据经济能力选择合适的缴费基数档次</li>
<li>办理参保登记：通过当地人社局官网或社保服务大厅办理灵活就业人员参保登记</li>
<li>按时缴费：每月按时足额缴纳社保费用，避免断缴影响待遇</li>
<li>定期查看权益：通过本平台跟踪缴费记录和权益状态</li>
</ol>

<h2>风险提示</h2>
<div style="background:#FEF3C7;border:1px solid #F59E0B;border-radius:6px;padding:10px 14px;margin:6px 0 14px">
<p style="color:#92400E;font-weight:600;margin-bottom:4px">请仔细阅读以下风险提示：</p>
<ul style="padding-left:18px;color:#78350F;font-size:12px;line-height:1.8">
<li><b>政策变动风险：</b>社保政策可能随国家和地方政策调整而变化，本报告基于当前有效政策生成，建议定期更新。</li>
<li><b>断缴风险：</b>社保断缴可能影响购房资格、落户积分、医保待遇等连续性要求，请确保按时足额缴纳。</li>
<li><b>基数选择风险：</b>缴费基数过低可能导致退休后养老金偏低，过高则增加当期经济压力，请根据个人经济能力合理选择。</li>
<li><b>不合规操作风险：</b>切勿通过虚构劳动关系、挂靠代缴等方式违规参保，可能导致行政处罚和个人信用受损。</li>
<li><b>地区差异风险：</b>不同城市社保政策差异较大，跨地区转移社保关系时可能存在待遇差异，请提前咨询当地社保经办机构。</li>
</ul>
</div>

<div class="footer">
<p>本报告由 AI社保智筹 生成，仅供参考。具体政策以当地人社局官方发布为准。</p>
<p>AI社保智筹 - 智能社保规划平台</p>
</div>

</body>
</html>`
