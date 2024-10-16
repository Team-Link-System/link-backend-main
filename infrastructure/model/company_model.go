package model

import "time"

// TODO 링크 프로젝트에서 쓸 회사 모델 - 얘는 먼저 db에 직접 넣어야함
// TODO 공공 api에서 먼저 56만건 넣고, 직접 해야함
type Company struct {
	ID                          uint   `gorm:"primaryKey"`
	Name                        string `gorm:"size:255" default:""` //회사이름
	BusinessRegistrationNumber  string `gorm:"size:255" default:""` //회사 사업자등록번호
	RepresentativeName          string `gorm:"size:255" default:""` //대표 이름
	RepresentativeEmail         string `gorm:"size:255" default:""` //대표 이메일
	RepresentativePhoneNumber   string `gorm:"size:255" default:""` //대표 전화번호
	RepresentativeAddress       string `gorm:"size:255" default:""` //대표 주소
	RepresentativeAddressDetail string `gorm:"size:255" default:""` //대표 주소 상세
	RepresentativePostalCode    string `gorm:"size:255" default:""` //대표 주소 우편번호
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
}
