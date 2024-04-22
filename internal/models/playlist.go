package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"math/rand"

	"gorm.io/gorm"
)

type JSONIntegerArray []int

func (j *JSONIntegerArray) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &j)
}

func (j JSONIntegerArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

type Playlist struct {
	gorm.Model
	Title         string  `gorm:"column:title;validate:min=10,max=200"`
	Owner         uint    `gorm:"column:owner_id;foreignKey:UserID"`
	Audios        []Audio `gorm:"many2many:playlist_audios;"`
	CoverURL      string  `gorm:"column:cover_url" validate:"omitempty,url"`
	CoverPublicID string  `gorm:"column:cover_public_id" validate:"omitempty,alphanum"`
	Visibility    string  `gorm:"column:visibility;default:public;validate:oneof=public private auto"`
}

func (p *Playlist) SetRandomCoverURL(db *gorm.DB) error {
	if len(p.Audios) > 0 {
		randomIndex := rand.Intn(len(p.Audios))
		p.CoverURL = p.Audios[randomIndex].CoverURL
	} else {
		p.CoverURL = "https://www.gstatic.com/youtube/media/ytm/images/pbg/playlist-empty-state-@576.png"
	}
	return nil
}
