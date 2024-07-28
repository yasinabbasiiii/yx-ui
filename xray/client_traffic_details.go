package xray

type ClientTrafficDetails struct {
	Id     int    `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Enable bool   `json:"enable" form:"enable"`
	Email  string `json:"email" form:"email" `
	Up     int64  `json:"up" form:"up"`
	Down   int64  `json:"down" form:"down"`
	Total  int64  `json:"total" form:"total"`
	Last   int64  `json:"last" form:"last"`
	Server string `json:"server" form:"server"`
}
