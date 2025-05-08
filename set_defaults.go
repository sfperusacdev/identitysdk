package identitysdk

var (
	identityAddress string
	accessToken     string
)

func SetIdentityServer(address string) { identityAddress = address }
func GetIdentityServer() string        { return identityAddress }

func SetAccessToken(token string) { accessToken = token }
func GetAccessToken() string      { return accessToken }
