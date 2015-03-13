package permission

type Permission interface {
	CreateGroup(name string) (groupId string, err error)
	DeleteGroup(groupId string) error
	AddUser(groupId string, email string) (permId string, err error)
	DeleteUser(groupId string, permId string) error
	GetUserList(groupId string) (emails []string, err error)
}
