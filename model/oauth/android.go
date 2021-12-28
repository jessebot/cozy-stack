package oauth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/cozy/cozy-stack/model/instance"
	"github.com/cozy/cozy-stack/pkg/config/config"
	jwt "github.com/golang-jwt/jwt/v4"
)

// checkAndroidAttestation will check an attestation made by the SafetyNet API.
// Cf https://developer.android.com/training/safetynet/attestation#use-response-server
func (c *Client) checkAndroidAttestation(inst *instance.Instance, attestation, challenge string) error {
	store := GetStore()
	if ok := store.CheckAndClearChallenge(inst, c.ID(), challenge); !ok {
		return errors.New("invalid challenge")
	}

	token, err := jwt.Parse(attestation, androidKeyFunc)
	if err != nil {
		return fmt.Errorf("cannot parse attestation: %s", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid claims type")
	}
	inst.Logger().Debugf("checkAndroidAttestation claims = %#v", claims)

	nonce, ok := claims["nonce"].(string)
	if !ok || len(nonce) == 0 {
		return errors.New("missing nonce")
	}
	if challenge != nonce {
		return errors.New("invalid nonce")
	}

	packageName, ok := claims["apkPackageName"].(string)
	if !ok || len(packageName) == 0 {
		return errors.New("missing apkPackageName")
	}
	names := config.GetConfig().Flagship.APKPackageNames
	for _, name := range names {
		if name == packageName {
			return nil
		}
	}
	return fmt.Errorf("%s is not the package name of the flagship app", packageName)
}

func androidKeyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	x5c, ok := token.Header["x5c"].([]interface{})
	if !ok || len(x5c) == 0 {
		return nil, errors.New("missing certification")
	}

	certs := make([]*x509.Certificate, 0, len(x5c))
	for _, raw := range x5c {
		rawStr, ok := raw.(string)
		if !ok {
			return nil, errors.New("missing certification")
		}
		buf, err := base64.StdEncoding.DecodeString(rawStr)
		if err != nil {
			return nil, fmt.Errorf("error decoding cert as base64: %s", err)
		}
		cert, err := x509.ParseCertificate(buf)
		if err != nil {
			return nil, fmt.Errorf("error parsing cert: %s", err)
		}
		certs = append(certs, cert)
	}
	intermediates := x509.NewCertPool()
	for _, cert := range certs {
		intermediates.AddCert(cert)
	}

	opts := x509.VerifyOptions{
		DNSName:       "attest.android.com",
		Intermediates: intermediates,
	}
	if _, err := certs[0].Verify(opts); err != nil {
		return nil, err
	}

	rsaKey, ok := certs[0].PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid certification")
	}
	return rsaKey, nil
}