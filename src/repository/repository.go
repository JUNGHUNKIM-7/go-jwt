package repository

type Sign struct {
	Email    string `bson:"email,omitempty" json:"email,omitempty"`
	Password string `bson:"password,omitempty" json:"password,omitempty"`
	Rt       string `bson:"rt,omitempty" json:"rt,omitempty"`
}