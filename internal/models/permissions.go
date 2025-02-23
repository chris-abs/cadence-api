package models

type Permission string

const (
	PermissionRead   Permission = "READ"
	PermissionWrite  Permission = "WRITE"
	PermissionManage Permission = "MANAGE"
)
