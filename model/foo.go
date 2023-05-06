package model

type ChatMsg struct {
	UserToken string `json:"userToken" gorm:"type:text;size:40;index"`
	ParentID  string `json:"parentID"`
	Prompt    string `json:"prompt"`
	NowID     string `json:"nowID"`
	Answer    string `json:"answer"`
}

type User struct {
	Token       string    `json:"token" gorm:"primaryKey"`
	Money       int       `json:"money"`
	ChatMsgInfo []ChatMsg `json:"chatMsgInfo" gorm:"foreignKey:UserToken"`
}
