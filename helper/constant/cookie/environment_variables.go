package constant

import "os"

const (
	JWT_S                     string = "JWT_S"
	MODE                      string = "MODE"
	SUPABASE_API_URL          string = "SUPABASE_API_URL"
	SUPABASE_SERVICE_ROLE_KEY string = "SUPABASE_SERVICE_ROLE_KEY"
	SUPABASE_PUBLIC_ANON_KEY  string = "SUPABASE_PUBLIC_ANON_KEY"
)

const MIDTRANS_SETTING string = "MIDTRANS_SETTING"

var MidtransEnvironment string = os.Getenv(MIDTRANS_SETTING)
