package models

import "time"

// User 用户模型(内存对齐)
type User struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"-"`
	UserID       int64      `gorm:"index:idx_user_id;unique;not null" json:"uid,string"`
	Age          int        `gorm:"column:age;type:int;check:age >= 0 AND age <= 150" json:"age"`
	Username     string     `gorm:"index:idx_username;type:varchar(64);not null" json:"username"`
	PasswordHash string     `gorm:"type:varchar(60);not null" json:"-"`
	Email        string     `gorm:"uniqueIndex:idx_email;type:varchar(320);unique" json:"email"`
	Gender       string     `gorm:"column:gender;default:male;type:ENUM('male','female') " json:"gender"`
	Signature    string     `gorm:"type:varchar(255);not null" json:"signature"`
	AvatarURL    string     `gorm:"type:varchar(255);default:''" json:"-"`
	Token        string     `gorm:"type:varchar(512);default:''" json:"token"`
	Birthday     *time.Time `gorm:"type:DATE" json:"birthday"`
	CreatedAt    time.Time  `gorm:"column:create_at;autoCreateTime" json:"created_at"`
	LastLogin    *time.Time `gorm:"column:last_login" json:"last_login"`
}

//type User struct {
//	ID           int64      `gorm:"primaryKey;autoIncrement" json:"-"`                                     //主键id
//	UserID       int64      `gorm:"index:idx_user_id;unique;not null" json:"uid,string"`                   //用户id
//	Username     string     `gorm:"index:idx_username;unique;type:varchar(64);not null" json:"username"`   //用户名
//	PasswordHash string     `gorm:"type:varchar(60);not null" json:"-"`                                    //密码hash
//	Email        string     `gorm:"uniqueIndex:idx_email;type:varchar(320)" json:"email"`                  //邮箱
//	Gender       string     `gorm:"column:gender;default:male;type:ENUM('male','female') '" json:"gender"` //性别
//	Signature    string     `gorm:"type:varchar(255);not null" json:"signature"`                           //个性签名
//	Age          int        `gorm:"column:age;type:int;check:age >= 0 AND age <= 150" json:"age" `         //年龄
//	AvatarURL    string     `gorm:"type:varchar(255);default:''" json:"-"`                                 //头像url
//	Birthday     *time.Time `gorm:"type:DATE" json:"birthday"`                                             //生日
//	CreatedAt    time.Time  `gorm:"column:create_at;autoCreateTime" json:"created_at"`                     //创建时间
//	LastLogin    *time.Time `gorm:"column:last_login" json:"last_login"`                                   //最后登录时间
//}
