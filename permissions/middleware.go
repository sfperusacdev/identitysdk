package permissions

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
	"github.com/user0608/goones/answer"
	"github.com/user0608/goones/errs"
)

type PermissionMiddlewareBuilder func(permissions []string) echo.MiddlewareFunc

func NewPermissionMiddlewareBuilder() PermissionMiddlewareBuilder {
	return func(permissions []string) echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				ctx := c.Request().Context()
				sucursal := identitysdk.Sucursal(ctx)

				if identitysdk.HasPermission(ctx, "admin", sucursal) {
					return next(c)
				}

				var missingPermissions []string
				for _, permission := range permissions {
					const globalPrefix = "g:"
					isGlobal := strings.HasPrefix(permission, globalPrefix)
					if isGlobal {
						permission = strings.TrimPrefix(permission, globalPrefix)
						if !identitysdk.HasPerm(ctx, permission) {
							missingPermissions = append(missingPermissions, permission)
						}
					} else {
						if !identitysdk.HasPermission(ctx, permission, sucursal) {
							missingPermissions = append(missingPermissions, permission)
						}
					}
				}

				if len(missingPermissions) > 0 {
					message := "No tienes permiso para realizar esta acción. Permisos requeridos: " + strings.Join(missingPermissions, ", ")
					return answer.Err(c, errs.ForbiddenDirect(message))
				}

				return next(c)
			}
		}
	}
}
