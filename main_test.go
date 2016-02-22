package main

import "testing"
import "log"
import "github.com/bmizerany/assert"
import "os"

var (
	cosgoAPIKeyTest      string
)

func TestTest(T *testing.T) {
  log.Println("\n\n\tcosgo v0.4\n\tCopyright 2016 aerth\n\tSource code at https://github.com/aerth/cosgo")
}

func TestRandom(t *testing.T) {
    log.Println("Test Random Key Generator")
    cosgoAPIKey = GenerateAPIKey(20)
    cosgoAPIKeyTest := GenerateAPIKey(20)
    assert.NotEqual(t, cosgoAPIKey, cosgoAPIKeyTest)
}
func TestReadEnv(t *testing.T) {
    log.Println("Test Environmental Variable get/set ")
    os.Setenv("COSGO_API_KEY", "123")
    key := os.Getenv("COSGO_API_KEY")
    keytest := "123"
    assert.Equal(t, key, keytest)
}
func TestQuickSelfTest(t *testing.T) {
    os.Setenv("COSGO_DESTINATION", "fake@email.com")
    os.Setenv("MANDRILL_KEY", "12345")
    QuickSelfTest()
}
