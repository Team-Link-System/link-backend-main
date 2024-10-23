package model

import "time"

type CompanyGrade int

const (
	CompanyGradeBasic CompanyGrade = iota + 1
	CompanyGradePro
)

type Company struct {
	ID                        uint         `gorm:"primaryKey"`
	CpName                    string       `json:"cp_name" gorm:"size:255;unique;not null" default:""`               //회사이름 TODO unique
	CpNumber                  string       `json:"cp_number,omitempty" gorm:"size:255;unique" default:""`            //회사 사업자등록번호 //unique null 허용
	RepresentativeName        string       `json:"representative_name,omitempty" gorm:"size:255" default:""`         //대표 이름
	RepresentativeEmail       string       `json:"representative_email,omitempty" gorm:"size:255" default:""`        //대표 이메일
	RepresentativePhoneNumber string       `json:"representative_phone_number,omitempty" gorm:"size:255" default:""` //대표 전화번호
	RepresentativeAddress     string       `json:"representative_address,omitempty" gorm:"size:255" default:""`      //대표 주소
	RepresentativePostalCode  string       `json:"representative_postal_code,omitempty" gorm:"size:255" default:""`  //대표 주소 우편번호
	IsVerified                bool         `json:"is_verified" gorm:"default:false"`                                 // 인증하게 되면 Basic 등급이 됨
	Grade                     CompanyGrade `json:"grade,omitempty" gorm:"default:null"`                              // 인증 받으면 Basic 등급이 됨
	Departments               []Department `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Teams                     []Team       `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	CreatedAt                 time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                 time.Time    `json:"updated_at"`
}
