package main

import (
	"log"
	"testing"
)

var (
	cosgoAPIKeyTest string
)

func TestTest(T *testing.T) {
	log.Println("\n\n\tcosgo v0.4\n\tCopyright 2016 aerth\n\tSource code at https://github.com/aerth/cosgo")
}

//
// func TestRandom(t *testing.T) {
// 	log.Println("Test Random Key Generator")
// 	cosgoAPIKey := generateAPIKey(20)
// 	cosgoAPIKeyTest := generateAPIKey(20)
// 	assert.NotEqual(t, cosgoAPIKey, cosgoAPIKeyTest)
// }
// func TestReadEnv(t *testing.T) {
// 	log.Println("Test Environmental Variable get/set ")
// 	os.Setenv("COSGO_API_KEY", "123")
// 	key := os.Getenv("COSGO_API_KEY")
// 	keytest := "123"
// 	assert.Equal(t, key, keytest)
// }
// func TestquickSelfTest(t *testing.T) {
// 	os.Setenv("COSGO_DESTINATION", "fake@email.com")
// 	os.Setenv("MANDRILL_KEY", "12345")
// 	quickSelfTest()
// }
//
// func TestOpenMailbox(t *testing.T) {
// 	mbox.Destination = "test@test.com"
//
//
// 	return true
// }
