package route

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
)

// TestCreateUserForm_Validate contains various test cases for the validate method.
func TestCreateUserForm_Validate(t *testing.T) {
	tests := []struct {
		name             string
		form             createUserForm
		expected         []string
		expectedUsername string // To check if username was lowercased
	}{
		{
			name: "Valid User",
			form: createUserForm{
				Username: "testuser",
				Password: "StrongPassword123",
				IsAdmin:  false,
			},
			expected:         []string{},
			expectedUsername: "testuser",
		},
		{
			name: "Username Too Short",
			form: createUserForm{
				Username: "ab", // Length 2, min is 3
				Password: "StrongPassword123",
				IsAdmin:  false,
			},
			expected:         []string{fmt.Sprintf("username must be between %d-%d characters long", minUsernameLength, maxUsernameLength)},
			expectedUsername: "ab",
		},
		{
			name: "Username Too Long",
			form: createUserForm{
				Username: "thisisareallylongusername12345", // Length 29, max is 20
				Password: "StrongPassword123",
				IsAdmin:  false,
			},
			expected:         []string{fmt.Sprintf("username must be between %d-%d characters long", minUsernameLength, maxUsernameLength)},
			expectedUsername: "thisisareallylongusername12345",
		},
		{
			name: "Username with Special Characters",
			form: createUserForm{
				Username: "user@name",
				Password: "StrongPassword123",
				IsAdmin:  false,
			},
			expected:         []string{"username must only contain alphanumeric characters"},
			expectedUsername: "user@name",
		},
		{
			name: "Username with Spaces",
			form: createUserForm{
				Username: "user name",
				Password: "StrongPassword123",
				IsAdmin:  false,
			},
			expected:         []string{"username must only contain alphanumeric characters"},
			expectedUsername: "user name",
		},
		{
			name: "Username with Uppercase - Should be Lowercased",
			form: createUserForm{
				Username: "TestUser",
				Password: "StrongPassword123",
				IsAdmin:  false,
			},
			expected:         []string{},
			expectedUsername: "testuser", // Expect it to be lowercased
		},
		{
			name: "Password Too Short",
			form: createUserForm{
				Username: "testuser",
				Password: "short", // Length 5, min is 8
				IsAdmin:  false,
			},
			expected:         []string{fmt.Sprintf("password must be between %d-%d characters long", minPasswordLength, maxPasswordLength)},
			expectedUsername: "testuser",
		},
		{
			name: "Password Too Long",
			form: createUserForm{
				Username: "testuser",
				Password: "thisisareallylongpasswordthatshouldbetoolong1234567890", // Length 54, max is 30
				IsAdmin:  false,
			},
			expected:         []string{fmt.Sprintf("password must be between %d-%d characters long", minPasswordLength, maxPasswordLength)},
			expectedUsername: "testuser",
		},
		{
			name: "Multiple Problems (Username Invalid, Password Too Short)",
			form: createUserForm{
				Username: "us@r",  // Invalid char
				Password: "short", // Too short
				IsAdmin:  false,
			},
			expected:         []string{"username must only contain alphanumeric characters", fmt.Sprintf("password must be between %d-%d characters long", minPasswordLength, maxPasswordLength)},
			expectedUsername: "us@r",
		},
		{
			name: "Multiple Problems (Username Too Short, Password Too Long)",
			form: createUserForm{
				Username: "a",                                                      // Too short
				Password: "thisisareallylongpasswordthatshouldbetoolong1234567890", // Too long
				IsAdmin:  false,
			},
			expected:         []string{fmt.Sprintf("username must be between %d-%d characters long", minUsernameLength, maxUsernameLength), fmt.Sprintf("password must be between %d-%d characters long", minPasswordLength, maxPasswordLength)},
			expectedUsername: "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the form for each test case to avoid side effects
			// since the validate method modifies the Username field.
			formToTest := tt.form

			problems := formToTest.validate(context.Background())

			// Check if the username was correctly lowercased
			if formToTest.Username != tt.expectedUsername {
				t.Errorf("For test '%s', username after validation was '%s', expected '%s'", tt.name, formToTest.Username, tt.expectedUsername)
			}

			// Check if the problems returned match the expected problems
			if !reflect.DeepEqual(problems, tt.expected) {
				t.Errorf("For test '%s', got problems %v, want %v", tt.name, problems, tt.expected)
			}
		})
	}
}

func Fuzz_createUserForm_validate(f *testing.F) {
	type fields struct {
		Username string
		Password string
		IsAdmin  bool
	}

	testcases := []fields{
		{
			Username: "user",
			Password: "password123",
			IsAdmin:  true,
		},
		{
			Username: "asdf",
			Password: "fdsa",
			IsAdmin:  false,
		},
		{
			Username: "w3ya0vwt0ywwa3900t39w",
			Password: "t70avy37r20qtvb3708q39",
			IsAdmin:  true,
		},
	}
	for _, c := range testcases {
		f.Add(c.Username, c.Password, c.IsAdmin)
	}
	f.Fuzz(func(t *testing.T, username, password string, isAdmin bool) {
		form := &createUserForm{
			Username: username,
			Password: password,
			IsAdmin:  isAdmin,
		}
		ch := make(chan struct{})
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Millisecond))
		go func(form *createUserForm, ctx context.Context, ch chan<- struct{}) {
			form.validate(ctx)
			ch <- struct{}{}
		}(form, ctx, ch)
		select {
		case <-ctx.Done():
			t.Errorf("Input took too long to complete: %v", form)
		case <-ch:
			break
		}
		cancel()
	})
}
