package plugins

import (
	"context"
	"io/fs"
	"testing"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersWithoutUtmpFile(t *testing.T) {
	orig := getUsers
	t.Cleanup(func() { getUsers = orig })

	getUsers = func(context.Context) ([]host.UserStat, error) {
		return nil, &fs.PathError{Op: "open", Path: "/var/run/utmp", Err: fs.ErrNotExist}
	}

	assert.Empty(t, Users(context.Background(), dynamotd.NoSettings{}))
}

func TestUsers(t *testing.T) {
	orig := getUsers
	t.Cleanup(func() { getUsers = orig })

	tests := []struct {
		name     string
		users    []string
		expected string
	}{
		{"nobody", nil, ""},
		{"one", []string{"alice"}, "alice"},
		{"many with many sessions", []string{"bob", "alice", "bob", "carol", "bob"}, "bob (3 sessions), alice, carol"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getUsers = func(context.Context) ([]host.UserStat, error) {
				var users []host.UserStat
				for _, name := range tt.users {
					users = append(users, host.UserStat{User: name, Terminal: "pts/0"})
				}
				return users, nil
			}

			rows := Users(context.Background(), dynamotd.NoSettings{})

			if tt.expected == "" {
				assert.Empty(t, rows)
				return
			}
			require.Len(t, rows, 1)
			assert.Equal(t, usersLabel, rows[0].Label.String())
			assert.Equal(t, tt.expected, rows[0].Value.String())
		})
	}
}
