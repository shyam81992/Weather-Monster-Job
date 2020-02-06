package models

// WebHook modal
type WebHook struct {
	ID          int64   `form:"id" json:"id"`
	CityID      int64   `form:"city_id" json:"city_id" binding:"required"`
	CallbackURL float64 `form:"callback_url" json:"callback_url" binding:"required"`
	Timestamp   int64   `form:"timestamp" json:"timestamp"`
}

// Temperature modal
type Temperature struct {
	//ID        int64   `form:"id" json:"id"`
	CityID    int64   `form:"city_id" json:"city_id"`
	Max       float64 `form:"max" json:"max"`
	Min       float64 `form:"min" json:"min"`
	Timestamp int64   `form:"timestamp"`
}
