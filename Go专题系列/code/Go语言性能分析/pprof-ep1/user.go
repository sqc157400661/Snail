package main

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	Telephone string `json:"telephone"`
	Introduce string `json:"introduce"`
	State     int    `json:"state"`
}

func getUser() User {
	return User{
		ID:        111,
		Username:  "裸奔的蜗牛",
		Password:  "111111111",
		Email:     "157400661@qq.com",
		Telephone: "1378888888",
		State:     1,
	}
}
