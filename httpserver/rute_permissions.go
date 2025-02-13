package httpapi

type PermissionChecker interface {
	CheckPermissions() []string
}
