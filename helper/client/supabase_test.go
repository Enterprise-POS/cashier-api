package client

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateSupabaseClient(t *testing.T) {
	client := CreateSupabaseClient()
	require.NotNil(t, client)
	require.Equal(t, "*supabase.Client", reflect.TypeOf(client).String())
}
