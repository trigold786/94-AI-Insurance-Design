package handler

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

func GuideHandler(evaluator *ComplianceEvaluator, policyRepo PolicyQuerier, profileRepo ProfileRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextKeyUserID).(string)
		if userID == "" {
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{"code": "UNAUTHORIZED", "message": "missing user"})
			return
		}

		profile, err := profileRepo.GetByUserID(r.Context(), userID)
		if err != nil {
			respondError(w, err)
			return
		}

		cityCode := r.URL.Query().Get("city_code")
		if cityCode == "" {
			cityCode = profile.CurrentResidenceCode
			if cityCode == "" {
				cityCode = profile.HouseholdRegionCode
			}
		}
		if cityCode == "" {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{"code": "VALIDATION_ERROR", "message": "city_code required"})
			return
		}

		policies, err := policyRepo.QueryByRegionAndStatus(r.Context(), cityCode, "verified")
		if err != nil {
			respondError(w, err)
			return
		}

		today := time.Now().Truncate(24 * time.Hour)

		var steps []models.ProcessingStep
		stepMap := make(map[int]bool)
		var allDocs []models.RequiredDocument
		docMap := make(map[string]bool)
		var matchedPolicies []models.PolicyCompliance

		for _, p := range policies {
			if p.EffectiveDate != "" {
				effectiveDate, parseErr := time.Parse("2006-01-02", p.EffectiveDate)
				if parseErr != nil {
					log.Printf("[guide] failed to parse effective_date for %s: %v", p.ClaimID, parseErr)
					continue
				}
				if effectiveDate.After(today) {
					continue
				}
			}
			if p.ExpireDate != nil && *p.ExpireDate != "" {
				expireDate, parseErr := time.Parse("2006-01-02", *p.ExpireDate)
				if parseErr == nil && expireDate.Before(today) {
					continue
				}
			}

			pc := models.PolicyCompliance{
				PolicyID:   p.PolicyID,
				PolicyType: p.PolicyType,
				ClaimID:    p.ClaimID,
			}
			if len(p.Conditions) > 0 {
				var conds []models.ComplianceCondition
				if err := jsonUnmarshalNoError(p.Conditions, &conds); err == nil {
					pc.Conditions = conds
				}
			}
			if len(p.RequiredDocuments) > 0 {
				var docs []models.RequiredDocument
				if err := jsonUnmarshalNoError(p.RequiredDocuments, &docs); err == nil {
					pc.RequiredDocs = docs
					for _, d := range docs {
						if !docMap[d.Name] {
							docMap[d.Name] = true
							allDocs = append(allDocs, d)
						}
					}
				}
			}
			isEligible, _ := evaluator.Evaluate(profile, &p)
			pc.IsEligible = isEligible
			pc.ProcessingSteps = getProcessingSteps(p.PolicyType, pc.Conditions, pc.RequiredDocs)
			for _, s := range pc.ProcessingSteps {
				if !stepMap[s.Order] {
					stepMap[s.Order] = true
					steps = append(steps, s)
				}
			}
			matchedPolicies = append(matchedPolicies, pc)
		}

		cityName := map[string]string{
			"110000": "北京", "310000": "上海", "440100": "广州",
			"330100": "杭州", "440300": "深圳", "310115": "上海浦东",
			"330108": "杭州滨江",
		}[cityCode]
		if cityName == "" {
			cityName = cityCode
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl := template.Must(template.New("guide").Parse(guideHTML))
		tmpl.Execute(w, map[string]interface{}{
			"CityCode":  cityCode,
			"CityName":  cityName,
			"Docs":      allDocs,
			"Steps":     steps,
			"Policies":  matchedPolicies,
		})
	})
}

func jsonUnmarshalNoError(data []byte, v interface{}) error {
	if data == nil || len(data) == 0 || string(data) == "null" {
		return nil
	}
	return json.Unmarshal(data, v)
}

const guideHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>办理指南 - AI社保智筹</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:"Microsoft YaHei","PingFang SC",sans-serif;color:#333;padding:20px;font-size:14px;line-height:1.6;background:#F9FAFB}
h1{font-size:20px;color:#1A56DB;margin-bottom:16px}
h2{font-size:16px;color:#1A56DB;border-bottom:2px solid #1A56DB;padding-bottom:6px;margin:20px 0 12px}
.flowchart{display:flex;flex-direction:column;align-items:center;gap:0;margin:16px 0}
.flow-node{background:#fff;border:2px solid #1A56DB;border-radius:10px;padding:10px 20px;text-align:center;min-width:200px;max-width:320px;box-shadow:0 2px 6px rgba(0,0,0,0.08);position:relative}
.flow-node.process{border-color:#1A56DB;background:#EFF6FF}
.flow-node.decision{border-color:#D97706;background:#FEF3C7;border-radius:40px;transform:rotate(0deg)}
.flow-node.success{border-color:#059669;background:#D1FAE5}
.flow-node .node-title{font-weight:600;font-size:14px;margin-bottom:2px}
.flow-node .node-desc{color:#6B7280;font-size:12px}
.flow-node .node-duration{display:inline-block;margin-top:4px;padding:1px 8px;border-radius:8px;font-size:11px;background:#E5E7EB;color:#6B7280}
.flow-arrow{width:2px;height:24px;background:#9CA3AF;margin:0 auto;position:relative}
.flow-arrow::after{content:"";position:absolute;bottom:0;left:50%;transform:translateX(-50%);border-left:5px solid transparent;border-right:5px solid transparent;border-top:8px solid #9CA3AF}
.flow-arrow.branch{height:32px}
.flow-branch-label{position:absolute;left:calc(50% + 12px);top:50%;transform:translateY(-50%);font-size:11px;color:#6B7280;background:#fff;padding:0 4px}
table{width:100%;border-collapse:collapse;margin:8px 0;background:#fff;border-radius:8px;overflow:hidden}
th{background:#F3F4F6;text-align:left;padding:8px 10px;border-bottom:2px solid #E5E7EB;color:#6B7280;font-size:12px}
td{padding:7px 10px;border-bottom:1px solid #F3F4F6;font-size:13px}
.badge{display:inline-block;padding:1px 6px;border-radius:4px;font-size:11px;font-weight:500}
.bg-green{background:#D1FAE5;color:#059669}
.bg-yellow{background:#FEF3C7;color:#D97706}
.card{background:#fff;border-radius:8px;padding:12px;margin:8px 0;box-shadow:0 1px 4px rgba(0,0,0,0.06)}
@media(max-width:600px){.flow-node{min-width:160px;max-width:260px}}
</style>
</head>
<body>

<h1>📋 办理指南 - {{.CityName}} ({{.CityCode}})</h1>

<h2>🔄 办理流程图</h2>
<div class="flowchart">
{{range .Steps}}
<div class="flow-node {{if eq .Order 1}}process{{else if eq .Order 2}}process{{else if eq .Name "审核公示"}}decision{{else if eq .Name "补贴发放"}}success{{else if eq .Name "待遇申领"}}success{{else}}process{{end}}">
<div class="node-title">{{.Order}}. {{.Name}}</div>
<div class="node-desc">{{.Description}}</div>
<div class="node-duration">⏱ {{if eq .Order 1}}约1-2天{{else if eq .Order 2}}即时提交{{else if eq .Order 3}}约5-15个工作日{{else if eq .Order 4}}约10个工作日{{else}}约1-7天{{end}}</div>
</div>
{{if not (eq .Order (len $.Steps))}}
<div class="flow-arrow"></div>
{{end}}
{{else}}
<div class="flow-node process"><div class="node-title">暂无流程数据</div></div>
{{end}}
</div>

<h2>📝 材料清单</h2>
<table>
<tr><th>材料名称</th><th>来源</th><th>说明</th></tr>
{{range .Docs}}
<tr><td>{{.Name}}</td><td>{{if eq .Source "user"}}🧑 个人准备{{else if eq .Source "gov"}}🏛 政府出具{{else}}🏢 单位提供{{end}}</td><td>{{.Description}}</td></tr>
{{else}}
<tr><td colspan="3" style="text-align:center;color:#9CA3AF">暂无材料要求</td></tr>
{{end}}
</table>

<h2>📌 匹配政策详情</h2>
{{range .Policies}}
<div class="card">
<div><strong>{{.PolicyType}}</strong> - {{.ClaimID}}</div>
{{if .IsEligible}}<div style="margin-top:4px"><span class="badge bg-green">✅ 符合条件</span></div>{{else}}<div style="margin-top:4px"><span class="badge bg-yellow">⚠️ 部分条件未满足</span></div>{{end}}
<div style="font-size:13px;color:#6B7280;margin-top:4px">计算方式: {{.SubsidyCalcMethod}}</div>
{{if .ProcessingSteps}}
<div style="font-size:12px;color:#9CA3AF;margin-top:6px">流程: {{range .ProcessingSteps}}{{.Name}} → {{end}}</div>
{{end}}
</div>
{{else}}
<div style="color:#9CA3AF;text-align:center;padding:20px">该城市暂无匹配政策</div>
{{end}}

<div style="border-top:1px solid #E5E7EB;text-align:center;color:#9CA3AF;font-size:12px;padding:12px 0;margin-top:20px">
<p>本指南由 AI社保智筹 生成，仅供参考。具体办理流程以当地人社局官方要求为准。</p>
</div>

</body>
</html>`

