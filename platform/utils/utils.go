package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"reflect"

	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func GenerateRandomString(length int, includeSpecial bool) string {
	str := letterBytes

	randString := make([]byte, length)
	_, _ = io.ReadAtLeast(rand.Reader, randString, length)
	for i := 0; i < len(randString); i++ {
		randString[i] = str[int(randString[i])%len(str)]
	}

	return string(randString)
}

func StructEqual(expected, actual interface{}, ignoredFields []string) error {
	expectedVal := reflect.ValueOf(expected)
	actualVal := reflect.ValueOf(actual)

	// Ensure both inputs are structs
	if expectedVal.Kind() == reflect.Ptr {
		expectedVal = expectedVal.Elem()
	}
	if actualVal.Kind() == reflect.Ptr {
		actualVal = actualVal.Elem()
	}

	if expectedVal.Kind() != reflect.Struct || actualVal.Kind() != reflect.Struct {
		return fmt.Errorf("both values must be structs")
	}

	// Convert ignored fields to a map for quick lookup
	ignoredMap := make(map[string]struct{}, len(ignoredFields))
	for _, field := range ignoredFields {
		ignoredMap[field] = struct{}{}
	}

	// Compare struct fields
	for i := 0; i < expectedVal.NumField(); i++ {
		field := expectedVal.Type().Field(i)
		fieldName := field.Name

		// Skip ignored fields
		if _, ok := ignoredMap[fieldName]; ok {
			continue
		}

		expectedField := expectedVal.Field(i)
		actualField := actualVal.FieldByName(fieldName)

		if !actualField.IsValid() {
			return fmt.Errorf("field %s is missing in actual struct", fieldName)
		}

		if !reflect.DeepEqual(expectedField.Interface(), actualField.Interface()) {
			return fmt.Errorf("field %s mismatch: expected %v, got %v",
				fieldName, expectedField.Interface(), actualField.Interface())
		}
	}

	return nil
}

const bcryptCost = 12

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func MakeTransactionWithBalance(
	ctx context.Context,
	wallet1, wallet2 decimal.Decimal,
	amount decimal.Decimal,
) (decimal.Decimal, decimal.Decimal, error) {
	// Make the transaction
	wallet := wallet1.Sub(amount)
	wallet2 = wallet2.Add(wallet)

	return wallet, wallet2, nil
}
