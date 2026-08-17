package auth


import "testing"


func TestCheckPasswordHashWrongPassword(t *testing.T) {
	truepassword := "helloworld"
	truehash, err := HashPassword(truepassword)
	if err != nil {
		t.Errorf("expected false, got error")
	}
	enteredpassword := "byeworld"
	match, err := CheckPasswordHash(enteredpassword, truehash)
	if err != nil {
		t.Errorf("expected false, got error")
	}
	if match {
		t.Errorf("expected false, got true")
	}
}


func TestCheckPasswordHashCorrectPassword(t *testing.T) {
	truepassword := "helloworld"
	truehash, err := HashPassword(truepassword)
	if err != nil {
		t.Errorf("expected true, got error")
	}
	match, err := CheckPasswordHash(truepassword, truehash)
	if err != nil {
		t.Errorf("expected true, got error")
	}
	if !match {
		t.Errorf("expected true, got false")
	}
}


