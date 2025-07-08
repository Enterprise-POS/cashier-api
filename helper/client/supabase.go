package client

import (
	"fmt"
	"os"

	"github.com/supabase-community/supabase-go"
)

var (
	API_URL          string = os.Getenv("SUPABASE_API_URL")
	SERVICE_ROLE_KEY string = os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
)

func CreateSupabaseClient() *supabase.Client {
	client, err := supabase.NewClient(API_URL, SERVICE_ROLE_KEY, &supabase.ClientOptions{})
	if err != nil {
		panic(fmt.Sprintf("cannot initalize client: %s", err.Error()))
	}

	return client
}
