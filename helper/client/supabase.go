package client

import (
	"fmt"
	"os"

	"github.com/supabase-community/supabase-go"
)

var (
	// MODE             string = os.Getenv("MODE")
	API_URL          string = os.Getenv("SUPABASE_API_URL")
	SERVICE_ROLE_KEY string = os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
)

func CreateSupabaseClient() *supabase.Client {
	// SERVICE_ROLE_KEY will bypass all the RLS. for dev it's allowed to use
	client, err := supabase.NewClient(API_URL, SERVICE_ROLE_KEY, &supabase.ClientOptions{})

	// When deleting user (not user from database) then this is required
	// For now use only for development, otherwise prevent use this
	// if MODE == "dev" {
	// 	session := types.Session{}
	// 	session.AccessToken = SERVICE_ROLE_KEY
	// 	client.UpdateAuthSession(session)
	// }

	if err != nil {
		panic(fmt.Sprintf("cannot initalize client: %s", err.Error()))
	}

	return client
}
