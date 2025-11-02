package models

var ModelFieldGetters = map[string]any{
	"User": map[string]func(*User)any{
		"id": func(u *User) any { return u.ID },
		"name": func(u *User) any { return u.Name },
		"email": func(u *User) any { return u.Email },
		"age": func(u *User) any { return u.Age },
		"is_active": func(u *User) any { return u.IsActive },
		"created_at": func(u *User) any { return u.CreatedAt },
		"friend": func(u *User) any { return u.Friend },
	},
	"UserFriend": map[string]func(*UserFriend)any{
		"id": func(u *UserFriend) any { return u.ID },
		"name": func(u *UserFriend) any { return u.Name },
	},
}
