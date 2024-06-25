package identitysdk

import "context"

func IsAdmin(ctx context.Context) bool { return HasPerm(ctx, "admin") }

func HasPerm(ctx context.Context, permission string) bool {
	session, ok := ReadSession(ctx)
	if !ok {
		return ok
	}
	for _, perm := range session.Permissions {
		if perm.ID == permission {
			return true
		}
	}
	return false
}

func HasPermission(ctx context.Context, permission, companyBranch string) bool {
	session, ok := ReadSession(ctx)
	if !ok {
		return ok
	}
	for _, perm := range session.Permissions {
		if perm.ID == permission {
			for _, b := range perm.CompanyBrances {
				if b == companyBranch {
					return true
				}
			}
		}
	}
	return false
}

func GetSubordinates(ctx context.Context) []string {
	session, ok := ReadSession(ctx)
	if !ok {
		return []string{}
	}
	return session.Subordinates
}

func GetSupervisors(ctx context.Context) []string {
	session, ok := ReadSession(ctx)
	if !ok {
		return []string{}
	}
	return session.Supervisors
}
