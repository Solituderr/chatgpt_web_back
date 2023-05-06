package model

import "fmt"

// 创建chatmsg
func CreateChatMsg(token string, msg ChatMsg) error {
	if err := DB.Create(&msg).Error; err != nil {
		return err
	}
	var newUser User
	newUser.Token = token
	err1 := DB.Model(&newUser).Association("ChatMsgInfo").Append(msg)
	if err1 != nil {
		return err1
	}
	return nil
}

// 查询是否存在id
func CheckExistId(token string, nowId string) bool {
	var cms []ChatMsg
	var user User
	DB.Where("token = ?", token).Find(&user)
	err := DB.Model(&user).Association("ChatMsgInfo").Find(&cms)
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < len(cms); i++ {
		if cms[i].NowID == nowId {
			return true
		}
	}
	return false
}

// 根据id查msg
func GetChatMsg(pid string) (ChatMsg, error) {
	var msg ChatMsg
	err := DB.Model(&ChatMsg{}).Where("parentID = ?", pid).First(&msg).Error
	return msg, err
}

// 根据token和money
func CreateUser(token string, money int) error {
	var user User
	user.Money = money
	user.Token = token
	err := DB.Create(&user).Error
	return err
}

// 扣款余额
func MinusToken(delMoney int, token string) error {
	var user User
	err := DB.Model(&User{}).Where("token = ?", token).Find(&user).Error
	if err != nil {
		return err
	}
	m := user.Money - delMoney
	err1 := DB.Model(&User{}).Where("token = ?", token).Update("money", m).Error
	if err1 != nil {
		return err1
	}
	return nil
}

// 扣除次数
func DeleteChatNum(token string) error {
	var user User
	err := DB.Model(&User{}).Where("token = ?", token).Find(&user).Error
	if err != nil {
		return err
	}
	m := user.Money - 1
	err1 := DB.Model(&User{}).Where("token = ?", token).Update("money", m).Error
	if err1 != nil {
		return err1
	}
	return nil
}

// 查询余额是否用完
func CheckMoney(token string) (string, error) {
	var user User
	err := DB.Where("token = ?", token).Find(&user).Error
	if err != nil {
		return "ok", err
	}
	if user.Money <= 0 {
		return "fail", nil
	}
	return "success", nil
}
