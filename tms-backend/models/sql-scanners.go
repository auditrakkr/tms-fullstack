package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Implement `sql.Scanner` for WebServerProperties
func (w *WebServerProperties) Scan(value interface{}) error {
    if value == nil {
        *w = WebServerProperties{}
        return nil
    }

    bytes, ok := value.([]byte)
    if !ok {
        return fmt.Errorf("failed to convert value to []byte")
    }

    if err := json.Unmarshal(bytes, w); err != nil {
        return fmt.Errorf("failed to unmarshal WebServerProperties: %w", err)
    }

    return nil
}

// Implement `driver.Valuer` for WebServerProperties
func (w WebServerProperties) Value() (driver.Value, error) {
    bytes, err := json.Marshal(w)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal WebServerProperties: %w", err)
    }

    return bytes, nil
}


// Implement `sql.Scanner` for DBProperties
func (d *DBProperties) Scan(value interface{}) error {
    if value == nil {
        *d = DBProperties{}
        return nil
    }

    bytes, ok := value.([]byte)
    if !ok {
        return fmt.Errorf("failed to convert value to []byte")
    }

    if err := json.Unmarshal(bytes, d); err != nil {
        return fmt.Errorf("failed to unmarshal DBProperties: %w", err)
    }

    return nil
}

// Implement `driver.Valuer` for DBProperties
func (d DBProperties) Value() (driver.Value, error) {
    bytes, err := json.Marshal(d)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal DBProperties: %w", err)
    }

    return bytes, nil
}


// Implement `sql.Scanner` for ElasticSearchProperties
func (e *ElasticSearchProperties) Scan(value interface{}) error {
    if value == nil {
        *e = ElasticSearchProperties{}
        return nil
    }

    bytes, ok := value.([]byte)
    if !ok {
        return fmt.Errorf("failed to convert value to []byte")
    }

    if err := json.Unmarshal(bytes, e); err != nil {
        return fmt.Errorf("failed to unmarshal ElasticSearchProperties: %w", err)
    }

    return nil
}

// Implement `driver.Valuer` for ElasticSearchProperties
func (e ElasticSearchProperties) Value() (driver.Value, error) {
    bytes, err := json.Marshal(e)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal ElasticSearchProperties: %w", err)
    }

    return bytes, nil
}

// Implement `sql.Scanner` for RedisProperties
func (r *RedisProperties) Scan(value interface{}) error {
    if value == nil {
        *r = RedisProperties{}
        return nil
    }

    bytes, ok := value.([]byte)
    if !ok {
        return fmt.Errorf("failed to convert value to []byte")
    }

    if err := json.Unmarshal(bytes, r); err != nil {
        return fmt.Errorf("failed to unmarshal RedisProperties: %w", err)
    }

    return nil
}

// Implement `driver.Valuer` for RedisProperties
func (r RedisProperties) Value() (driver.Value, error) {
    bytes, err := json.Marshal(r)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal RedisProperties: %w", err)
    }

    return bytes, nil
}

// Implement `sql.Scanner` for RootFileSystem
func (r *RootFileSystem) Scan(value interface{}) error {
	if value == nil {
		*r = RootFileSystem{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert value to []byte")
	}

	if err := json.Unmarshal(bytes, r); err != nil {
		return fmt.Errorf("failed to unmarshal RootFileSystem: %w", err)
	}

	return nil
}

// Implement `driver.Valuer` for RootFileSystem
func (r RootFileSystem) Value() (driver.Value, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RootFileSystem: %w", err)
	}

	return bytes, nil
}
// Implement `sql.Scanner` for SMTPAuth
func (s *SMTPAuth) Scan(value interface{}) error {
	if value == nil {
		*s = SMTPAuth{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert value to []byte")
	}

	if err := json.Unmarshal(bytes, s); err != nil {
		return fmt.Errorf("failed to unmarshal SMTPAuth: %w", err)
	}

	return nil
}

// Implement `driver.Valuer` for SMTPAuth
func (s SMTPAuth) Value() (driver.Value, error) {
	bytes, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SMTPAuth: %w", err)
	}

	return bytes, nil
}

// Implement `sql.Scanner` for JWTConstants
func (j *JWTConstants) Scan(value interface{}) error {
	if value == nil {
		*j = JWTConstants{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert value to []byte")
	}

	if err := json.Unmarshal(bytes, j); err != nil {
		return fmt.Errorf("failed to unmarshal JWTConstants: %w", err)
	}

	return nil
}

// Implement `driver.Valuer` for JWTConstants
func (j JWTConstants) Value() (driver.Value, error) {
	bytes, err := json.Marshal(j)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JWTConstants: %w", err)
	}

	return bytes, nil
}

// Implement `sql.Scanner` for AuthEnabled
func (a *AuthEnabled) Scan(value interface{}) error {
	if value == nil {
		*a = AuthEnabled{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert value to []byte")
	}

	if err := json.Unmarshal(bytes, a); err != nil {
		return fmt.Errorf("failed to unmarshal AuthEnabled: %w", err)
	}

	return nil
}

// Implement `driver.Valuer` for AuthEnabled
func (a AuthEnabled) Value() (driver.Value, error) {
	bytes, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AuthEnabled: %w", err)
	}

	return bytes, nil
}

// Implement `sql.Scanner` for FBOauth2Constants
func (f *FBOauth2Constants) Scan(value interface{}) error {
	if value == nil {
		*f = FBOauth2Constants{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert value to []byte")
	}

	if err := json.Unmarshal(bytes, f); err != nil {
		return fmt.Errorf("failed to unmarshal FBOauth2Constants: %w", err)
	}

	return nil
}
// Implement `driver.Valuer` for FBOauth2Constants
func (f FBOauth2Constants) Value() (driver.Value, error) {
	bytes, err := json.Marshal(f)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal FBOauth2Constants: %w", err)
	}

	return bytes, nil
}
// Implement `sql.Scanner` for GoogleOauth2Constants
func (g *GoogleOauth2Constants) Scan(value interface{}) error {
	if value == nil {
		*g = GoogleOauth2Constants{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert value to []byte")
	}

	if err := json.Unmarshal(bytes, g); err != nil {
		return fmt.Errorf("failed to unmarshal GoogleOauth2Constants: %w", err)
	}

	return nil
}

// Implement `driver.Valuer` for GoogleOauth2Constants
func (g GoogleOauth2Constants) Value() (driver.Value, error) {
	bytes, err := json.Marshal(g)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GoogleOauth2Constants: %w", err)
	}

	return bytes, nil
}

// Implement `sql.Scanner` for OtherUserOptions
func (o *OtherUserOptions) Scan(value interface{}) error {
	if value == nil {
		*o = OtherUserOptions{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert value to []byte")
	}

	if err := json.Unmarshal(bytes, o); err != nil {
		return fmt.Errorf("failed to unmarshal OtherUserOptions: %w", err)
	}

	return nil
}

// Implement `driver.Valuer` for OtherUserOptions
func (o OtherUserOptions) Value() (driver.Value, error) {
	bytes, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OtherUserOptions: %w", err)
	}

	return bytes, nil
}

// Implement `sql.Scanner` for SizeLimits
func (s *SizeLimits) Scan(value interface{}) error {
	if value == nil {
		*s = SizeLimits{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert value to []byte")
	}

	if err := json.Unmarshal(bytes, s); err != nil {
		return fmt.Errorf("failed to unmarshal SizeLimits: %w", err)
	}

	return nil
}

// Implement `driver.Valuer` for SizeLimits
func (s SizeLimits) Value() (driver.Value, error) {
	bytes, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SizeLimits: %w", err)
	}

	return bytes, nil
}

// Implement `sql.Scanner` for ThemeType
func (t *ThemeType) Scan(value interface{}) error {
	if value == nil {
		*t = ThemeType{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert value to []byte")
	}

	if err := json.Unmarshal(bytes, t); err != nil {
		return fmt.Errorf("failed to unmarshal ThemeType: %w", err)
	}

	return nil
}
// Implement `driver.Valuer` for ThemeType
func (t ThemeType) Value() (driver.Value, error) {
	bytes, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ThemeType: %w", err)
	}

	return bytes, nil
}

// Implement `sql.Scanner` for Logo
func (l *Logo) Scan(value interface{}) error {
	if value == nil {
		*l = Logo{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert value to []byte")
	}

	if err := json.Unmarshal(bytes, l); err != nil {
		return fmt.Errorf("failed to unmarshal Logo: %w", err)
	}

	return nil
}
// Implement `driver.Valuer` for Logo
func (l Logo) Value() (driver.Value, error) {
	bytes, err := json.Marshal(l)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Logo: %w", err)
	}

	return bytes, nil
}