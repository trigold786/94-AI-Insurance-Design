package admin

import (
	"encoding/json"
	"net/http"
)

type Region struct {
	Code   string   `json:"code"`
	Name   string   `json:"name"`
	Level  string   `json:"level"`
	Parent string   `json:"parent,omitempty"`
}

type RegionResponse struct {
	Provinces []Region  `json:"provinces"`
	Cities    []Region  `json:"cities"`
	Districts []Region  `json:"districts"`
}

var chinaRegions = []Region{
	{Code: "110000", Name: "北京市", Level: "province"},
	{Code: "110100", Name: "北京市辖区", Level: "city", Parent: "110000"},
	{Code: "110101", Name: "东城区", Level: "district", Parent: "110100"},
	{Code: "110102", Name: "西城区", Level: "district", Parent: "110100"},
	{Code: "110105", Name: "朝阳区", Level: "district", Parent: "110100"},
	{Code: "110106", Name: "丰台区", Level: "district", Parent: "110100"},
	{Code: "110107", Name: "石景山区", Level: "district", Parent: "110100"},
	{Code: "110108", Name: "海淀区", Level: "district", Parent: "110100"},
	{Code: "110109", Name: "门头沟区", Level: "district", Parent: "110100"},
	{Code: "110111", Name: "房山区", Level: "district", Parent: "110100"},
	{Code: "110112", Name: "通州区", Level: "district", Parent: "110100"},
	{Code: "110113", Name: "顺义区", Level: "district", Parent: "110100"},
	{Code: "110114", Name: "昌平区", Level: "district", Parent: "110100"},
	{Code: "110115", Name: "大兴区", Level: "district", Parent: "110100"},

	// 北京市街道级
	{Code: "110101001", Name: "东华门街道", Level: "street", Parent: "110101"}, // 东城区
	{Code: "110101002", Name: "景山街道", Level: "street", Parent: "110101"},
	{Code: "110101003", Name: "交道口街道", Level: "street", Parent: "110101"},
	{Code: "110108001", Name: "中关村街道", Level: "street", Parent: "110108"}, // 海淀区
	{Code: "110108002", Name: "学院路街道", Level: "street", Parent: "110108"},
	{Code: "110105001", Name: "建外街道", Level: "street", Parent: "110105"}, // 朝阳区
	{Code: "110105002", Name: "三里屯街道", Level: "street", Parent: "110105"},

	{Code: "120000", Name: "天津市", Level: "province"},
	{Code: "120100", Name: "天津市辖区", Level: "city", Parent: "120000"},
	{Code: "120101", Name: "和平区", Level: "district", Parent: "120100"},
	{Code: "120102", Name: "河东区", Level: "district", Parent: "120100"},
	{Code: "120103", Name: "河西区", Level: "district", Parent: "120100"},
	{Code: "120104", Name: "南开区", Level: "district", Parent: "120100"},
	{Code: "120105", Name: "河北区", Level: "district", Parent: "120100"},

	{Code: "130000", Name: "河北省", Level: "province"},
	{Code: "130100", Name: "石家庄市", Level: "city", Parent: "130000"},
	{Code: "130200", Name: "唐山市", Level: "city", Parent: "130000"},
	{Code: "130300", Name: "秦皇岛市", Level: "city", Parent: "130000"},
	{Code: "130400", Name: "邯郸市", Level: "city", Parent: "130000"},
	{Code: "130500", Name: "邢台市", Level: "city", Parent: "130000"},
	{Code: "130600", Name: "保定市", Level: "city", Parent: "130000"},
	{Code: "130700", Name: "张家口市", Level: "city", Parent: "130000"},
	{Code: "130800", Name: "承德市", Level: "city", Parent: "130000"},

	{Code: "140000", Name: "山西省", Level: "province"},
	{Code: "140100", Name: "太原市", Level: "city", Parent: "140000"},
	{Code: "140200", Name: "大同市", Level: "city", Parent: "140000"},

	{Code: "150000", Name: "内蒙古自治区", Level: "province"},
	{Code: "150100", Name: "呼和浩特市", Level: "city", Parent: "150000"},
	{Code: "150200", Name: "包头市", Level: "city", Parent: "150000"},

	{Code: "210000", Name: "辽宁省", Level: "province"},
	{Code: "210100", Name: "沈阳市", Level: "city", Parent: "210000"},
	{Code: "210200", Name: "大连市", Level: "city", Parent: "210000"},

	{Code: "220000", Name: "吉林省", Level: "province"},
	{Code: "220100", Name: "长春市", Level: "city", Parent: "220000"},

	{Code: "230000", Name: "黑龙江省", Level: "province"},
	{Code: "230100", Name: "哈尔滨市", Level: "city", Parent: "230000"},
	{Code: "230200", Name: "齐齐哈尔市", Level: "city", Parent: "230000"},

	{Code: "310000", Name: "上海市", Level: "province"},
	{Code: "310100", Name: "上海市辖区", Level: "city", Parent: "310000"},
	{Code: "310101", Name: "黄浦区", Level: "district", Parent: "310100"},
	{Code: "310104", Name: "徐汇区", Level: "district", Parent: "310100"},
	{Code: "310105", Name: "长宁区", Level: "district", Parent: "310100"},
	{Code: "310106", Name: "静安区", Level: "district", Parent: "310100"},
	{Code: "310107", Name: "普陀区", Level: "district", Parent: "310100"},
	{Code: "310109", Name: "虹口区", Level: "district", Parent: "310100"},
	{Code: "310110", Name: "杨浦区", Level: "district", Parent: "310100"},
	{Code: "310112", Name: "闵行区", Level: "district", Parent: "310100"},
	{Code: "310113", Name: "宝山区", Level: "district", Parent: "310100"},
	{Code: "310114", Name: "嘉定区", Level: "district", Parent: "310100"},
	{Code: "310115", Name: "浦东新区", Level: "district", Parent: "310100"},
	{Code: "310116", Name: "金山区", Level: "district", Parent: "310100"},
	{Code: "310117", Name: "松江区", Level: "district", Parent: "310100"},
	{Code: "310118", Name: "青浦区", Level: "district", Parent: "310100"},
	{Code: "310120", Name: "奉贤区", Level: "district", Parent: "310100"},

	// 上海市街道级
	{Code: "310115001", Name: "花木街道", Level: "street", Parent: "310115"}, // 浦东新区
	{Code: "310115002", Name: "张江镇", Level: "street", Parent: "310115"},
	{Code: "310115003", Name: "陆家嘴街道", Level: "street", Parent: "310115"},
	{Code: "310101001", Name: "南京东路街道", Level: "street", Parent: "310101"}, // 黄浦区
	{Code: "310101002", Name: "外滩街道", Level: "street", Parent: "310101"},
	{Code: "310106001", Name: "静安寺街道", Level: "street", Parent: "310106"}, // 静安区
	{Code: "310106002", Name: "南京西路街道", Level: "street", Parent: "310106"},

	{Code: "320000", Name: "江苏省", Level: "province"},
	{Code: "320100", Name: "南京市", Level: "city", Parent: "320000"},
	{Code: "320200", Name: "无锡市", Level: "city", Parent: "320000"},
	{Code: "320300", Name: "徐州市", Level: "city", Parent: "320000"},
	{Code: "320400", Name: "常州市", Level: "city", Parent: "320000"},
	{Code: "320500", Name: "苏州市", Level: "city", Parent: "320000"},
	{Code: "320600", Name: "南通市", Level: "city", Parent: "320000"},
	{Code: "320700", Name: "连云港市", Level: "city", Parent: "320000"},
	{Code: "320800", Name: "淮安市", Level: "city", Parent: "320000"},
	{Code: "320900", Name: "盐城市", Level: "city", Parent: "320000"},
	{Code: "321000", Name: "扬州市", Level: "city", Parent: "320000"},
	{Code: "321100", Name: "镇江市", Level: "city", Parent: "320000"},
	{Code: "321200", Name: "泰州市", Level: "city", Parent: "320000"},
	{Code: "321300", Name: "宿迁市", Level: "city", Parent: "320000"},

	{Code: "330000", Name: "浙江省", Level: "province"},
	{Code: "330100", Name: "杭州市", Level: "city", Parent: "330000"},
	{Code: "330102", Name: "上城区", Level: "district", Parent: "330100"},
	{Code: "330103", Name: "下城区", Level: "district", Parent: "330100"},
	{Code: "330106", Name: "西湖区", Level: "district", Parent: "330100"},
	{Code: "330108", Name: "滨江区", Level: "district", Parent: "330100"},
	{Code: "330109", Name: "萧山区", Level: "district", Parent: "330100"},
	{Code: "330110", Name: "余杭区", Level: "district", Parent: "330100"},

	// 杭州市街道级
	{Code: "330106001", Name: "北山街道", Level: "street", Parent: "330106"}, // 西湖区
	{Code: "330106002", Name: "翠苑街道", Level: "street", Parent: "330106"},
	{Code: "330106003", Name: "西溪街道", Level: "street", Parent: "330106"},
	{Code: "330102001", Name: "湖滨街道", Level: "street", Parent: "330102"}, // 上城区
	{Code: "330102002", Name: "清波街道", Level: "street", Parent: "330102"},
	{Code: "330108001", Name: "浦沿街道", Level: "street", Parent: "330108"}, // 滨江区
	{Code: "330108002", Name: "西兴街道", Level: "street", Parent: "330108"},

	{Code: "330200", Name: "宁波市", Level: "city", Parent: "330000"},
	{Code: "330300", Name: "温州市", Level: "city", Parent: "330000"},
	{Code: "330400", Name: "嘉兴市", Level: "city", Parent: "330000"},
	{Code: "330500", Name: "湖州市", Level: "city", Parent: "330000"},
	{Code: "330600", Name: "绍兴市", Level: "city", Parent: "330000"},

	{Code: "340000", Name: "安徽省", Level: "province"},
	{Code: "340100", Name: "合肥市", Level: "city", Parent: "340000"},

	{Code: "350000", Name: "福建省", Level: "province"},
	{Code: "350100", Name: "福州市", Level: "city", Parent: "350000"},
	{Code: "350200", Name: "厦门市", Level: "city", Parent: "350000"},

	{Code: "360000", Name: "江西省", Level: "province"},
	{Code: "360100", Name: "南昌市", Level: "city", Parent: "360000"},

	{Code: "370000", Name: "山东省", Level: "province"},
	{Code: "370100", Name: "济南市", Level: "city", Parent: "370000"},
	{Code: "370200", Name: "青岛市", Level: "city", Parent: "370000"},

	{Code: "410000", Name: "河南省", Level: "province"},
	{Code: "410100", Name: "郑州市", Level: "city", Parent: "410000"},

	{Code: "420000", Name: "湖北省", Level: "province"},
	{Code: "420100", Name: "武汉市", Level: "city", Parent: "420000"},

	{Code: "430000", Name: "湖南省", Level: "province"},
	{Code: "430100", Name: "长沙市", Level: "city", Parent: "430000"},

	{Code: "440000", Name: "广东省", Level: "province"},
	{Code: "440100", Name: "广州市", Level: "city", Parent: "440000"},
	{Code: "440103", Name: "荔湾区", Level: "district", Parent: "440100"},
	{Code: "440104", Name: "越秀区", Level: "district", Parent: "440100"},
	{Code: "440105", Name: "海珠区", Level: "district", Parent: "440100"},
	{Code: "440106", Name: "天河区", Level: "district", Parent: "440100"},
	{Code: "440111", Name: "白云区", Level: "district", Parent: "440100"},
	{Code: "440112", Name: "黄埔区", Level: "district", Parent: "440100"},
	{Code: "440113", Name: "番禺区", Level: "district", Parent: "440100"},
	{Code: "440114", Name: "花都区", Level: "district", Parent: "440100"},
	{Code: "440115", Name: "南沙区", Level: "district", Parent: "440100"},

	// 广州市街道级
	{Code: "440106001", Name: "五山街道", Level: "street", Parent: "440106"}, // 天河区
	{Code: "440106002", Name: "天园街道", Level: "street", Parent: "440106"},
	{Code: "440106003", Name: "员村街道", Level: "street", Parent: "440106"},
	{Code: "440104001", Name: "北京街道", Level: "street", Parent: "440104"}, // 越秀区
	{Code: "440104002", Name: "六榕街道", Level: "street", Parent: "440104"},
	{Code: "440105001", Name: "赤岗街道", Level: "street", Parent: "440105"}, // 海珠区
	{Code: "440105002", Name: "新港街道", Level: "street", Parent: "440105"},

	{Code: "440200", Name: "韶关市", Level: "city", Parent: "440000"},
	{Code: "440300", Name: "深圳市", Level: "city", Parent: "440000"},
	{Code: "440303", Name: "罗湖区", Level: "district", Parent: "440300"},
	{Code: "440304", Name: "福田区", Level: "district", Parent: "440300"},
	{Code: "440305", Name: "南山区", Level: "district", Parent: "440300"},
	{Code: "440306", Name: "宝安区", Level: "district", Parent: "440300"},
	{Code: "440307", Name: "龙岗区", Level: "district", Parent: "440300"},
	{Code: "440308", Name: "盐田区", Level: "district", Parent: "440300"},

	// 深圳市街道级
	{Code: "440304001", Name: "福田街道", Level: "street", Parent: "440304"}, // 福田区
	{Code: "440304002", Name: "香蜜湖街道", Level: "street", Parent: "440304"},
	{Code: "440305001", Name: "南头街道", Level: "street", Parent: "440305"}, // 南山区
	{Code: "440305002", Name: "蛇口街道", Level: "street", Parent: "440305"},
	{Code: "440303001", Name: "桂园街道", Level: "street", Parent: "440303"}, // 罗湖区
	{Code: "440303002", Name: "东门街道", Level: "street", Parent: "440303"},

	{Code: "440400", Name: "珠海市", Level: "city", Parent: "440000"},
	{Code: "440500", Name: "汕头市", Level: "city", Parent: "440000"},
	{Code: "440600", Name: "佛山市", Level: "city", Parent: "440000"},
	{Code: "440700", Name: "江门市", Level: "city", Parent: "440000"},
	{Code: "440800", Name: "湛江市", Level: "city", Parent: "440000"},

	{Code: "450000", Name: "广西壮族自治区", Level: "province"},
	{Code: "450100", Name: "南宁市", Level: "city", Parent: "450000"},

	{Code: "460000", Name: "海南省", Level: "province"},
	{Code: "460100", Name: "海口市", Level: "city", Parent: "460000"},
	{Code: "460200", Name: "三亚市", Level: "city", Parent: "460000"},

	{Code: "500000", Name: "重庆市", Level: "province"},
	{Code: "500100", Name: "重庆市辖区", Level: "city", Parent: "500000"},

	{Code: "510000", Name: "四川省", Level: "province"},
	{Code: "510100", Name: "成都市", Level: "city", Parent: "510000"},

	{Code: "520000", Name: "贵州省", Level: "province"},
	{Code: "520100", Name: "贵阳市", Level: "city", Parent: "520000"},

	{Code: "530000", Name: "云南省", Level: "province"},
	{Code: "530100", Name: "昆明市", Level: "city", Parent: "530000"},

	{Code: "540000", Name: "西藏自治区", Level: "province"},
	{Code: "540100", Name: "拉萨市", Level: "city", Parent: "540000"},

	{Code: "610000", Name: "陕西省", Level: "province"},
	{Code: "610100", Name: "西安市", Level: "city", Parent: "610000"},

	{Code: "620000", Name: "甘肃省", Level: "province"},
	{Code: "620100", Name: "兰州市", Level: "city", Parent: "620000"},

	{Code: "630000", Name: "青海省", Level: "province"},
	{Code: "630100", Name: "西宁市", Level: "city", Parent: "630000"},

	{Code: "640000", Name: "宁夏回族自治区", Level: "province"},
	{Code: "640100", Name: "银川市", Level: "city", Parent: "640000"},

	{Code: "650000", Name: "新疆维吾尔自治区", Level: "province"},
	{Code: "650100", Name: "乌鲁木齐市", Level: "city", Parent: "650000"},

	{Code: "000000", Name: "全国通用", Level: "province"},
}

func RegionsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parent := r.URL.Query().Get("parent")
		if parent != "" {
			var children []Region
			for _, reg := range chinaRegions {
				if reg.Parent == parent {
					children = append(children, reg)
				}
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": children})
			return
		}

		var provinces []Region
		for _, reg := range chinaRegions {
			if reg.Level == "province" {
				provinces = append(provinces, reg)
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": RegionResponse{
			Provinces: provinces,
		}})
	})
}
