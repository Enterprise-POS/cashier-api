package client

import (
	"fmt"

	"github.com/supabase-community/supabase-go"
)

func CreateSupabaseClient() *supabase.Client {
	client, err := supabase.NewClient(API_URL, SERVICE_ROLE_KEY, &supabase.ClientOptions{})
	if err != nil {
		panic(fmt.Sprintf("cannot initalize client: %s", err.Error()))
	}

	return client
}
