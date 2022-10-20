// Filename: internal/data/schools.go

package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"appletree.osborncollins.net/internal/validator"
	"github.com/lib/pq"
)

type School struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Name      string    `json:"name"`
	Level     string    `json:"level"`
	Contact   string    `json:"contact"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email,omitempty"`
	Website   string    `json:"website,omitempty"`
	Address   string    `json:"address"`
	Mode      []string  `json:"mode"`
	Version   int32     `json:"version"`
}

func ValidateSchool(v *validator.Validator, school *School) {

	// Use the check() method to execute our validation checks
	//Name
	v.Check(school.Name != "", "name", "must be provided")
	v.Check(len(school.Name) <= 200, "name", "must not be move than 200 bytes long")

	//Level
	v.Check(school.Level != "", "level", "must be provided")
	v.Check(len(school.Level) <= 200, "level", "must not be move than 200 bytes long")

	//Contact
	v.Check(school.Contact != "", "contact", "must be provided")
	v.Check(len(school.Contact) <= 200, "contact", "must not be move than 200 bytes long")

	//Phone
	v.Check(school.Phone != "", "phone", "must be provided")
	v.Check(validator.Matches(school.Phone, validator.PhoneRx), "phone", "must be a valid phone number")

	//Email
	v.Check(school.Email != "", "email", "must be provided")
	v.Check(validator.Matches(school.Phone, validator.EmailRx), "email", "must be a valid email address")

	//Website
	v.Check(school.Website != "", "website", "must be provided")
	v.Check(validator.ValidWebsite(school.Website), "website", "must be a valid URL")

	//Address
	//Contact
	v.Check(school.Address != "", "address", "must be provided")
	v.Check(len(school.Address) <= 500, "address", "must not be move than 500 bytes long")

	//Mode
	//Contact
	v.Check(school.Mode != nil, "mode", "must be provided")
	v.Check(len(school.Mode) >= 1, "mode", "must contain atleast 1 entry")
	v.Check(len(school.Mode) <= 5, "mode", "must contain atleast 5 entries")
	v.Check(validator.Unique(school.Mode), "mode", "must not contain duplicate entries")
}

// Define a SchoolModel which wraps a sql.DB connection pool
type SchoolModel struct {
	DB *sql.DB
}

// Insert() allows us to create a new school
func (m SchoolModel) Insert(school *School) error {
	query := `
	INSERT INTO schools (name, level, contact, phone, email, website, address, mode)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id, created_at, version
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Collect the data fields into a slice
	args := []interface{}{school.Name, school.Level, school.Contact, school.Phone,
		school.Email, school.Website, school.Address, pq.Array(school.Mode),
	}
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&school.ID, &school.CreatedAt, &school.Version)
}

// GET() allows us to retrieve a specific school
func (m SchoolModel) Get(id int64) (*School, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Create query
	query := `
		SELECT id, created_at, name, level, contact, phone, email, 
		website, address, mode, version
		FROM schools
		WHERE id = $1
	`
	// Declare a School variable to hold the return data
	var school School
	// Execute Query using the QueryRow
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&school.ID,
		&school.CreatedAt,
		&school.Name,
		&school.Level,
		&school.Contact,
		&school.Phone,
		&school.Email,
		&school.Website,
		&school.Address,
		pq.Array(&school.Mode),
		&school.Version,
	)
	// Handle any errors
	if err != nil {
		// Check the type of error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Success
	return &school, nil
}

// Update() allows us to edit/alter a specific school
func (m SchoolModel) Update(school *School) error {
	query := `
		UPDATE schools 
		set name = $1, level = $2, 
		contact = $3, phone = $4, 
		email = $5, website = $6, 
		address = $7, mode = $8, 
		version = version + 1
		WHERE id = $9
		AND version = $10
		RETURNING version
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()

	args := []interface{}{
		school.Name,
		school.Level,
		school.Contact,
		school.Phone,
		school.Email,
		school.Website,
		school.Address,
		pq.Array(school.Mode),
		school.ID,
		school.Version,
	}
	// Check for edit conflicts
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&school.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Delete() removes a specific school
func (m SchoolModel) Delete(id int64) error {
	// Ensure that there is a valid id
	if id < 1 {
		return ErrRecordNotFound
	}
	// Create the delete query
	query := `
		DELETE FROM schools
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Execute the query
	results, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// Check how many rows were affected by the delete operations. We
	// call the RowsAffected() method on the result variable
	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return err
	}
	// Check if no rows were affected
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// The GetAll() returns a list of all the school sorted by ID
func (m SchoolModel) GetAll(name string, level string, mode []string, filters Filters) ([]*School, Metadata, error) {
	// Construct the query
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, created_at, name, level, contact, phone, email, website, 
		address, mode, version
		FROM schools
		WHERE (to_tsvector('simple',name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (to_tsvector('simple',level) @@ plainto_tsquery('simple', $2) OR $2 = '')
		AND (mode @> $3 OR $3 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortOrder())

	// Create a 3-second-timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []interface{}{name, level, pq.Array(mode), filters.limit(), filters.offset()}
	// Execute query
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	// Close the result set
	defer rows.Close()
	totalRecords := 0
	// Initialize an empty slice to hold the school data
	schools := []*School{}
	// Iterate over the rows in the results set
	for rows.Next() {
		var school School
		// Scan the values from the row in to the School struct
		err := rows.Scan(
			&totalRecords,
			&school.ID,
			&school.Contact,
			&school.Name,
			&school.Level,
			&school.Contact,
			&school.Phone,
			&school.Email,
			&school.Website,
			&school.Address,
			pq.Array(&school.Mode),
			&school.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// Add the School to our slice
		schools = append(schools, &school)
	}
	// Check for errors after looping through the results set
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	// Return the slice of schools
	return schools, metadata, nil
}
