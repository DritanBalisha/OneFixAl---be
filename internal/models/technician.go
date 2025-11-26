package models

import "gorm.io/gorm"

type TechnicianProfile struct {
	gorm.Model
	UserID            uint    `json:"userId"`
	Professions       string  `json:"profession"` // easier: store as CSV string
	Bio               string  `json:"bio"`
	ProfilePicture    string  `json:"profile_picture"` // URL or file path
	Certificate       string  `json:"certificate"`     // could be file path
	YearsOfExperience string  `json:"experience"`
	RatingAvg         float32 `json:"ratingAvg"`
	Verified          bool    `json:"verified"`
}
