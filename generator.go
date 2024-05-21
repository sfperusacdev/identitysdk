package identitysdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func PermissionGoFileConstants() string {
	res, err := http.DefaultClient.Get(identityAddress)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	var result struct {
		Data []string `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Fatalln(err)
	}

	var builder strings.Builder
	builder.WriteString("package constants\n\n")
	builder.WriteString("type PermissionKey string\n\n")
	for _, value := range result.Data {
		if strings.Contains(value, " ") {
			log.Println("Warning: empty spaces found in key:", value)
		}
		key := strings.ToUpper(strings.ReplaceAll(value, ".", "_"))
		constValue := fmt.Sprintf(`const %s = PermissionKey("%s")`, key, value)
		builder.WriteString(fmt.Sprintln(constValue))
	}
	return builder.String()
}
