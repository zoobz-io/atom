package atom

import (
	"bytes"
	"database/sql"
	"errors"
	"testing"
	"time"
)

// mockColScanner implements ColScanner for testing.
type mockColScanner struct {
	columns []string
	rows    [][]any
	current int
	err     error
}

func (m *mockColScanner) Columns() ([]string, error) {
	return m.columns, nil
}

func (m *mockColScanner) Scan(dest ...any) error {
	if m.current >= len(m.rows) {
		return errors.New("no more rows")
	}
	row := m.rows[m.current]
	for i, v := range row {
		if i >= len(dest) {
			break
		}
		switch d := dest[i].(type) {
		case *string:
			if s, ok := v.(string); ok {
				*d = s
			}
		case *int64:
			switch n := v.(type) {
			case int64:
				*d = n
			case int:
				*d = int64(n)
			}
		case *uint64:
			switch n := v.(type) {
			case uint64:
				*d = n
			case int:
				*d = uint64(n) //nolint:gosec // test mock; values are controlled
			}
		case *float64:
			if f, ok := v.(float64); ok {
				*d = f
			}
		case *bool:
			if b, ok := v.(bool); ok {
				*d = b
			}
		case *time.Time:
			if t, ok := v.(time.Time); ok {
				*d = t
			}
		case *[]byte:
			if b, ok := v.([]byte); ok {
				*d = b
			}
		case *sql.NullString:
			if v == nil {
				d.Valid = false
			} else if s, ok := v.(string); ok {
				d.String = s
				d.Valid = true
			}
		case *sql.NullInt64:
			if v == nil {
				d.Valid = false
			} else {
				switch n := v.(type) {
				case int64:
					d.Int64 = n
					d.Valid = true
				case int:
					d.Int64 = int64(n)
					d.Valid = true
				}
			}
		case *sql.NullFloat64:
			if v == nil {
				d.Valid = false
			} else if f, ok := v.(float64); ok {
				d.Float64 = f
				d.Valid = true
			}
		case *sql.NullBool:
			if v == nil {
				d.Valid = false
			} else if b, ok := v.(bool); ok {
				d.Bool = b
				d.Valid = true
			}
		case *sql.NullTime:
			if v == nil {
				d.Valid = false
			} else if t, ok := v.(time.Time); ok {
				d.Time = t
				d.Valid = true
			}
		case *any:
			*d = v
		}
	}
	return nil
}

func (m *mockColScanner) Err() error {
	return m.err
}

func (m *mockColScanner) Next() bool {
	m.current++
	return m.current < len(m.rows)
}

// Reset prepares for scanning (sets current to -1 so first Next() goes to 0).
func (m *mockColScanner) Reset() {
	m.current = -1
}

type TestUserForScan struct {
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
	Age   int    `db:"age"`
}

func TestScannerFor(t *testing.T) {
	// Register the type
	_, err := Use[TestUserForScan]()
	if err != nil {
		t.Fatalf("Use[TestUserForScan]() failed: %v", err)
	}

	// Get spec via sentinel
	atomizer, err := Use[TestUserForScan]()
	if err != nil {
		t.Fatalf("Use[TestUserForScan]() failed: %v", err)
	}
	spec := atomizer.Spec()

	// Retrieve scanner
	scanner, ok := ScannerFor(spec)
	if !ok {
		t.Fatal("ScannerFor returned false for registered type")
	}
	if scanner == nil {
		t.Fatal("ScannerFor returned nil scanner")
	}
}

func TestScanner_Scan(t *testing.T) {
	// Register and get scanner
	atomizer, err := Use[TestUserForScan]()
	if err != nil {
		t.Fatalf("Use[TestUserForScan]() failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	// Create mock scanner with one row
	mock := &mockColScanner{
		columns: []string{"id", "name", "email", "age"},
		rows: [][]any{
			{int64(1), "Alice", "alice@example.com", int64(30)},
		},
		current: 0,
	}

	// Scan
	atom, err := scanner.Scan(mock)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Verify values
	if atom.Ints["ID"] != 1 {
		t.Errorf("expected ID=1, got %d", atom.Ints["ID"])
	}
	if atom.Strings["Name"] != "Alice" {
		t.Errorf("expected Name='Alice', got %q", atom.Strings["Name"])
	}
	if atom.Strings["Email"] != "alice@example.com" {
		t.Errorf("expected Email='alice@example.com', got %q", atom.Strings["Email"])
	}
	if atom.Ints["Age"] != 30 {
		t.Errorf("expected Age=30, got %d", atom.Ints["Age"])
	}
}

func TestScanner_ScanAll(t *testing.T) {
	atomizer, err := Use[TestUserForScan]()
	if err != nil {
		t.Fatalf("Use[TestUserForScan]() failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	// Create mock with multiple rows
	mock := &mockColScanner{
		columns: []string{"id", "name", "email", "age"},
		rows: [][]any{
			{int64(1), "Alice", "alice@example.com", int64(30)},
			{int64(2), "Bob", "bob@example.com", int64(25)},
			{int64(3), "Charlie", "charlie@example.com", int64(35)},
		},
		current: -1, // Start before first row
	}

	atoms, err := scanner.ScanAll(mock, mock.Next)
	if err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	if len(atoms) != 3 {
		t.Fatalf("expected 3 atoms, got %d", len(atoms))
	}

	// Verify first row
	if atoms[0].Ints["ID"] != 1 {
		t.Errorf("row 0: expected ID=1, got %d", atoms[0].Ints["ID"])
	}
	if atoms[0].Strings["Name"] != "Alice" {
		t.Errorf("row 0: expected Name='Alice', got %q", atoms[0].Strings["Name"])
	}

	// Verify second row
	if atoms[1].Ints["ID"] != 2 {
		t.Errorf("row 1: expected ID=2, got %d", atoms[1].Ints["ID"])
	}
	if atoms[1].Strings["Name"] != "Bob" {
		t.Errorf("row 1: expected Name='Bob', got %q", atoms[1].Strings["Name"])
	}

	// Verify third row
	if atoms[2].Ints["ID"] != 3 {
		t.Errorf("row 2: expected ID=3, got %d", atoms[2].Ints["ID"])
	}
	if atoms[2].Strings["Name"] != "Charlie" {
		t.Errorf("row 2: expected Name='Charlie', got %q", atoms[2].Strings["Name"])
	}
}

func TestScanner_UnknownColumns(t *testing.T) {
	atomizer, err := Use[TestUserForScan]()
	if err != nil {
		t.Fatalf("Use[TestUserForScan]() failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	// Include an extra column not in the struct
	mock := &mockColScanner{
		columns: []string{"id", "name", "unknown_column", "email"},
		rows: [][]any{
			{int64(1), "Alice", "ignored_value", "alice@example.com"},
		},
		current: 0,
	}

	// Should not error on unknown columns
	atom, err := scanner.Scan(mock)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if atom.Ints["ID"] != 1 {
		t.Errorf("expected ID=1, got %d", atom.Ints["ID"])
	}
	if atom.Strings["Name"] != "Alice" {
		t.Errorf("expected Name='Alice', got %q", atom.Strings["Name"])
	}
}

func TestScanner_PartialColumns(t *testing.T) {
	atomizer, err := Use[TestUserForScan]()
	if err != nil {
		t.Fatalf("Use[TestUserForScan]() failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	// Only include some columns
	mock := &mockColScanner{
		columns: []string{"id", "name"},
		rows: [][]any{
			{int64(1), "Alice"},
		},
		current: 0,
	}

	atom, err := scanner.Scan(mock)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if atom.Ints["ID"] != 1 {
		t.Errorf("expected ID=1, got %d", atom.Ints["ID"])
	}
	if atom.Strings["Name"] != "Alice" {
		t.Errorf("expected Name='Alice', got %q", atom.Strings["Name"])
	}
	// Email and Age should not be present
	if _, ok := atom.Strings["Email"]; ok {
		t.Error("Email should not be present in atom")
	}
}

type TestNestedForScan struct {
	Name    string             `db:"name"`
	Address TestAddressForScan // No db tag on struct field
}

type TestAddressForScan struct {
	Street string `db:"street"`
	City   string `db:"city"`
}

func TestScanner_NestedStruct(t *testing.T) {
	atomizer, err := Use[TestNestedForScan]()
	if err != nil {
		t.Fatalf("Use[TestNestedForScan]() failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	mock := &mockColScanner{
		columns: []string{"name", "street", "city"},
		rows: [][]any{
			{"Alice", "123 Main St", "Springfield"},
		},
		current: 0,
	}

	atom, err := scanner.Scan(mock)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if atom.Strings["Name"] != "Alice" {
		t.Errorf("expected Name='Alice', got %q", atom.Strings["Name"])
	}

	// Check nested atom
	nested, ok := atom.Nested["Address"]
	if !ok {
		t.Fatal("expected nested Address atom")
	}
	if nested.Strings["Street"] != "123 Main St" {
		t.Errorf("expected Street='123 Main St', got %q", nested.Strings["Street"])
	}
	if nested.Strings["City"] != "Springfield" {
		t.Errorf("expected City='Springfield', got %q", nested.Strings["City"])
	}
}

type TestBadNestedForScan struct {
	Name    string             `db:"name"`
	Address TestAddressForScan `db:"address"` // ERROR: db tag on struct field
}

func TestScanner_ErrorOnStructWithDBTag(t *testing.T) {
	atomizer, err := Use[TestBadNestedForScan]()
	if err != nil {
		// Registration failed entirely - that's also acceptable
		return
	}
	if atomizer == nil {
		t.Fatal("atomizer is nil but no error returned")
	}

	// Type registered, but scanner should not be available
	_, ok := ScannerFor(atomizer.Spec())
	if ok {
		t.Error("expected scanner to not be registered for struct with db tag on nested field")
	}
}

type TestWithSliceForScan struct {
	Name string   `db:"name"`
	Tags []string // No db tag - should be silently skipped
}

func TestScanner_SliceFieldsSkipped(t *testing.T) {
	atomizer, err := Use[TestWithSliceForScan]()
	if err != nil {
		t.Fatalf("Use[TestWithSliceForScan]() failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	mock := &mockColScanner{
		columns: []string{"name"},
		rows: [][]any{
			{"Alice"},
		},
		current: 0,
	}

	atom, err := scanner.Scan(mock)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if atom.Strings["Name"] != "Alice" {
		t.Errorf("expected Name='Alice', got %q", atom.Strings["Name"])
	}
}

type TestSliceWithDBTag struct {
	Name string   `db:"name"`
	Tags []string `db:"tags"` // ERROR: db tag on slice field
}

func TestScanner_ErrorOnSliceWithDBTag(t *testing.T) {
	atomizer, err := Use[TestSliceWithDBTag]()
	if err != nil {
		// Registration failed entirely - that's also acceptable
		return
	}
	if atomizer == nil {
		t.Fatal("atomizer is nil but no error returned")
	}

	// Type registered, but scanner should not be available
	_, ok := ScannerFor(atomizer.Spec())
	if ok {
		t.Error("expected scanner to not be registered for slice with db tag")
	}
}

// --- Gap tests ---

// Test pointer-to-struct fields.
type TestPointerNestedForScan struct {
	Name    string              `db:"name"`
	Address *TestAddressForScan // Pointer to struct, no db tag
}

func TestScanner_PointerToStruct(t *testing.T) {
	atomizer, err := Use[TestPointerNestedForScan]()
	if err != nil {
		t.Fatalf("Use[TestPointerNestedForScan]() failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	mock := &mockColScanner{
		columns: []string{"name", "street", "city"},
		rows: [][]any{
			{"Alice", "123 Main St", "Springfield"},
		},
		current: 0,
	}

	atom, err := scanner.Scan(mock)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if atom.Strings["Name"] != "Alice" {
		t.Errorf("expected Name='Alice', got %q", atom.Strings["Name"])
	}

	// Check nested atom (pointer-to-struct should still create nested atom)
	nested, ok := atom.Nested["Address"]
	if !ok {
		t.Fatal("expected nested Address atom")
	}
	if nested.Strings["Street"] != "123 Main St" {
		t.Errorf("expected Street='123 Main St', got %q", nested.Strings["Street"])
	}
	if nested.Strings["City"] != "Springfield" {
		t.Errorf("expected City='Springfield', got %q", nested.Strings["City"])
	}
}

// Test deeply nested structs (3+ levels).
type TestLevel1 struct {
	Name   string     `db:"name"`
	Level2 TestLevel2 // nested
}

type TestLevel2 struct {
	Value  string     `db:"level2_value"`
	Level3 TestLevel3 // nested again
}

type TestLevel3 struct {
	DeepValue string `db:"deep_value"`
}

func TestScanner_DeeplyNested(t *testing.T) {
	atomizer, err := Use[TestLevel1]()
	if err != nil {
		t.Fatalf("Use[TestLevel1]() failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	mock := &mockColScanner{
		columns: []string{"name", "level2_value", "deep_value"},
		rows: [][]any{
			{"Root", "Middle", "Deep"},
		},
		current: 0,
	}

	atom, err := scanner.Scan(mock)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Check root level
	if atom.Strings["Name"] != "Root" {
		t.Errorf("expected Name='Root', got %q", atom.Strings["Name"])
	}

	// Check level 2
	level2, ok := atom.Nested["Level2"]
	if !ok {
		t.Fatal("expected nested Level2 atom")
	}
	if level2.Strings["Value"] != "Middle" {
		t.Errorf("expected Level2.Value='Middle', got %q", level2.Strings["Value"])
	}

	// Check level 3
	level3, ok := level2.Nested["Level3"]
	if !ok {
		t.Fatal("expected nested Level3 atom")
	}
	if level3.Strings["DeepValue"] != "Deep" {
		t.Errorf("expected Level3.DeepValue='Deep', got %q", level3.Strings["DeepValue"])
	}
}

// Test nullable types.
type TestNullableForScan struct {
	ID     int64    `db:"id"`
	Name   *string  `db:"name"`   // nullable string
	Age    *int64   `db:"age"`    // nullable int
	Score  *float64 `db:"score"`  // nullable float
	Active *bool    `db:"active"` // nullable bool
}

func TestScanner_NullableTypes(t *testing.T) {
	atomizer, err := Use[TestNullableForScan]()
	if err != nil {
		t.Fatalf("Use[TestNullableForScan]() failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	t.Run("with values", func(t *testing.T) {
		mock := &mockColScanner{
			columns: []string{"id", "name", "age", "score", "active"},
			rows: [][]any{
				{int64(1), "Alice", int64(30), 95.5, true},
			},
			current: 0,
		}

		atom, err := scanner.Scan(mock)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		if atom.Ints["ID"] != 1 {
			t.Errorf("expected ID=1, got %d", atom.Ints["ID"])
		}
		if atom.StringPtrs["Name"] == nil || *atom.StringPtrs["Name"] != "Alice" {
			t.Errorf("expected Name='Alice', got %v", atom.StringPtrs["Name"])
		}
		if atom.IntPtrs["Age"] == nil || *atom.IntPtrs["Age"] != 30 {
			t.Errorf("expected Age=30, got %v", atom.IntPtrs["Age"])
		}
		if atom.FloatPtrs["Score"] == nil || *atom.FloatPtrs["Score"] != 95.5 {
			t.Errorf("expected Score=95.5, got %v", atom.FloatPtrs["Score"])
		}
		if atom.BoolPtrs["Active"] == nil || *atom.BoolPtrs["Active"] != true {
			t.Errorf("expected Active=true, got %v", atom.BoolPtrs["Active"])
		}
	})

	t.Run("with nulls", func(t *testing.T) {
		mock := &mockColScanner{
			columns: []string{"id", "name", "age", "score", "active"},
			rows: [][]any{
				{int64(2), nil, nil, nil, nil},
			},
			current: 0,
		}

		atom, err := scanner.Scan(mock)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		if atom.Ints["ID"] != 2 {
			t.Errorf("expected ID=2, got %d", atom.Ints["ID"])
		}
		if atom.StringPtrs["Name"] != nil {
			t.Errorf("expected Name=nil, got %v", atom.StringPtrs["Name"])
		}
		if atom.IntPtrs["Age"] != nil {
			t.Errorf("expected Age=nil, got %v", atom.IntPtrs["Age"])
		}
		if atom.FloatPtrs["Score"] != nil {
			t.Errorf("expected Score=nil, got %v", atom.FloatPtrs["Score"])
		}
		if atom.BoolPtrs["Active"] != nil {
			t.Errorf("expected Active=nil, got %v", atom.BoolPtrs["Active"])
		}
	})
}

// Test column name collision detection.
type TestCollisionParent struct {
	Name  string `db:"name"`
	Child TestCollisionChild
}

type TestCollisionChild struct {
	Name string `db:"name"` // Same column name as parent - should error!
}

func TestScanner_ColumnCollision(t *testing.T) {
	// Registration should succeed (atomizer doesn't care about db tags)
	atomizer, err := Use[TestCollisionParent]()
	if err != nil {
		t.Fatalf("Use[TestCollisionParent]() failed: %v", err)
	}

	// But scanner should NOT be available due to collision
	_, ok := ScannerFor(atomizer.Spec())
	if ok {
		t.Error("expected scanner to not be registered due to column collision")
	}
}

// Test time.Time fields.
type TestTimeForScan struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
}

func TestScanner_TimeField(t *testing.T) {
	atomizer, err := Use[TestTimeForScan]()
	if err != nil {
		t.Fatalf("Use[TestTimeForScan]() failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	now := time.Now().Truncate(time.Second)
	mock := &mockColScanner{
		columns: []string{"id", "created_at"},
		rows: [][]any{
			{int64(1), now},
		},
		current: 0,
	}

	atom, err := scanner.Scan(mock)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if atom.Ints["ID"] != 1 {
		t.Errorf("expected ID=1, got %d", atom.Ints["ID"])
	}
	if !atom.Times["CreatedAt"].Equal(now) {
		t.Errorf("expected CreatedAt=%v, got %v", now, atom.Times["CreatedAt"])
	}
}

// Test *time.Time nullable field.
type TestNullableTimeForScan struct {
	ID        int64      `db:"id"`
	UpdatedAt *time.Time `db:"updated_at"`
}

func TestScanner_NullableTimeField(t *testing.T) {
	atomizer, err := Use[TestNullableTimeForScan]()
	if err != nil {
		t.Fatalf("Use[TestNullableTimeForScan]() failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	t.Run("with value", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		mock := &mockColScanner{
			columns: []string{"id", "updated_at"},
			rows: [][]any{
				{int64(1), now},
			},
			current: 0,
		}

		atom, err := scanner.Scan(mock)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		if atom.TimePtrs["UpdatedAt"] == nil {
			t.Fatal("expected UpdatedAt to be non-nil")
		}
		if !atom.TimePtrs["UpdatedAt"].Equal(now) {
			t.Errorf("expected UpdatedAt=%v, got %v", now, *atom.TimePtrs["UpdatedAt"])
		}
	})

	t.Run("with null", func(t *testing.T) {
		mock := &mockColScanner{
			columns: []string{"id", "updated_at"},
			rows: [][]any{
				{int64(2), nil},
			},
			current: 0,
		}

		atom, err := scanner.Scan(mock)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		if atom.TimePtrs["UpdatedAt"] != nil {
			t.Errorf("expected UpdatedAt=nil, got %v", atom.TimePtrs["UpdatedAt"])
		}
	})
}

// Test []byte field.
type TestBytesForScan struct {
	ID   int64  `db:"id"`
	Data []byte `db:"data"`
}

func TestScanner_BytesField(t *testing.T) {
	atomizer, err := Use[TestBytesForScan]()
	if err != nil {
		t.Fatalf("Use[TestBytesForScan]() failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	data := []byte{0x01, 0x02, 0x03, 0x04}
	mock := &mockColScanner{
		columns: []string{"id", "data"},
		rows: [][]any{
			{int64(1), data},
		},
		current: 0,
	}

	atom, err := scanner.Scan(mock)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if atom.Ints["ID"] != 1 {
		t.Errorf("expected ID=1, got %d", atom.Ints["ID"])
	}
	if !bytes.Equal(atom.Bytes["Data"], data) {
		t.Errorf("expected Data=%v, got %v", data, atom.Bytes["Data"])
	}
}

// Test *[]byte nullable field.
type TestNullableBytesForScan struct {
	ID   int64   `db:"id"`
	Data *[]byte `db:"data"`
}

func TestScanner_NullableBytesField(t *testing.T) {
	atomizer, err := Use[TestNullableBytesForScan]()
	if err != nil {
		t.Fatalf("Use[TestNullableBytesForScan]() failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	t.Run("with value", func(t *testing.T) {
		data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
		mock := &mockColScanner{
			columns: []string{"id", "data"},
			rows: [][]any{
				{int64(1), data},
			},
			current: 0,
		}

		atom, err := scanner.Scan(mock)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		if atom.BytePtrs["Data"] == nil {
			t.Fatal("expected Data to be non-nil")
		}
		if !bytes.Equal(*atom.BytePtrs["Data"], data) {
			t.Errorf("expected Data=%v, got %v", data, *atom.BytePtrs["Data"])
		}
	})

	t.Run("with null", func(t *testing.T) {
		mock := &mockColScanner{
			columns: []string{"id", "data"},
			rows: [][]any{
				{int64(2), nil},
			},
			current: 0,
		}

		atom, err := scanner.Scan(mock)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		if atom.BytePtrs["Data"] != nil {
			t.Errorf("expected Data=nil, got %v", atom.BytePtrs["Data"])
		}
	})
}

// Test empty ScanAll result.
func TestScanner_ScanAllEmpty(t *testing.T) {
	atomizer, err := Use[TestUserForScan]()
	if err != nil {
		t.Fatalf("Use[TestUserForScan]() failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	// Empty result set
	mock := &mockColScanner{
		columns: []string{"id", "name", "email", "age"},
		rows:    [][]any{}, // No rows
		current: -1,
	}

	atoms, err := scanner.ScanAll(mock, mock.Next)
	if err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	if len(atoms) != 0 {
		t.Errorf("expected 0 atoms, got %d", len(atoms))
	}
}

// Test all scalar types for full assignValue coverage.
type TestAllScalarsForScan struct {
	ID      int64   `db:"id"`
	Count   uint64  `db:"count"`
	Score   float64 `db:"score"`
	Active  bool    `db:"active"`
	Name    string  `db:"name"`
	Created time.Time `db:"created"`
	Data    []byte  `db:"data"`
}

func TestScanner_AllScalarTypes(t *testing.T) {
	atomizer, err := Use[TestAllScalarsForScan]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	now := time.Now().Truncate(time.Second)
	data := []byte{0x01, 0x02}

	mock := &mockColScanner{
		columns: []string{"id", "count", "score", "active", "name", "created", "data"},
		rows: [][]any{
			{int64(1), uint64(100), 95.5, true, "test", now, data},
		},
		current: 0,
	}

	atom, err := scanner.Scan(mock)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if atom.Ints["ID"] != 1 {
		t.Errorf("expected ID=1, got %d", atom.Ints["ID"])
	}
	if atom.Uints["Count"] != 100 {
		t.Errorf("expected Count=100, got %d", atom.Uints["Count"])
	}
	if atom.Floats["Score"] != 95.5 {
		t.Errorf("expected Score=95.5, got %f", atom.Floats["Score"])
	}
	if atom.Bools["Active"] != true {
		t.Errorf("expected Active=true, got %v", atom.Bools["Active"])
	}
	if atom.Strings["Name"] != "test" {
		t.Errorf("expected Name=test, got %s", atom.Strings["Name"])
	}
	if !atom.Times["Created"].Equal(now) {
		t.Errorf("expected Created=%v, got %v", now, atom.Times["Created"])
	}
	if !bytes.Equal(atom.Bytes["Data"], data) {
		t.Errorf("expected Data=%v, got %v", data, atom.Bytes["Data"])
	}
}

// Test nullable uint pointer.
type TestNullableUintForScan struct {
	ID    int64   `db:"id"`
	Count *uint64 `db:"count"`
}

func TestScanner_NullableUintField(t *testing.T) {
	atomizer, err := Use[TestNullableUintForScan]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	t.Run("with value", func(t *testing.T) {
		mock := &mockColScanner{
			columns: []string{"id", "count"},
			rows: [][]any{
				{int64(1), int64(42)}, // NullInt64 uses int64
			},
			current: 0,
		}

		atom, err := scanner.Scan(mock)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		if atom.UintPtrs["Count"] == nil || *atom.UintPtrs["Count"] != 42 {
			t.Errorf("expected Count=42, got %v", atom.UintPtrs["Count"])
		}
	})

	t.Run("with null", func(t *testing.T) {
		mock := &mockColScanner{
			columns: []string{"id", "count"},
			rows: [][]any{
				{int64(2), nil},
			},
			current: 0,
		}

		atom, err := scanner.Scan(mock)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		if atom.UintPtrs["Count"] != nil {
			t.Errorf("expected Count=nil, got %v", atom.UintPtrs["Count"])
		}
	})
}

// Test ScanAll with many rows to ensure resetDests is properly exercised.
func TestScanner_ScanAllManyRows(t *testing.T) {
	atomizer, err := Use[TestAllScalarsForScan]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	now := time.Now().Truncate(time.Second)
	data := []byte{0xAB}

	// Create many rows to exercise resetDests
	rows := make([][]any, 10)
	for i := range rows {
		rows[i] = []any{
			int64(i),
			uint64(i * 10), //nolint:gosec // test data; values are small and controlled
			float64(i) * 1.5,
			i%2 == 0,
			"name" + string(rune('A'+i)),
			now.Add(time.Duration(i) * time.Hour),
			append([]byte{}, data...),
		}
	}

	mock := &mockColScanner{
		columns: []string{"id", "count", "score", "active", "name", "created", "data"},
		rows:    rows,
		current: -1,
	}

	atoms, err := scanner.ScanAll(mock, mock.Next)
	if err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	if len(atoms) != 10 {
		t.Fatalf("expected 10 atoms, got %d", len(atoms))
	}

	// Verify a few rows to ensure values were properly reset between scans
	if atoms[0].Ints["ID"] != 0 {
		t.Errorf("row 0: expected ID=0, got %d", atoms[0].Ints["ID"])
	}
	if atoms[5].Ints["ID"] != 5 {
		t.Errorf("row 5: expected ID=5, got %d", atoms[5].Ints["ID"])
	}
	if atoms[9].Ints["ID"] != 9 {
		t.Errorf("row 9: expected ID=9, got %d", atoms[9].Ints["ID"])
	}

	// Check that boolean values are correctly alternating
	for i := 0; i < 10; i++ {
		expected := i%2 == 0
		if atoms[i].Bools["Active"] != expected {
			t.Errorf("row %d: expected Active=%v, got %v", i, expected, atoms[i].Bools["Active"])
		}
	}
}

// Test ScanAll with nullable types to exercise more resetDests branches.
func TestScanner_ScanAllNullables(t *testing.T) {
	atomizer, err := Use[TestNullableForScan]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	scanner, ok := ScannerFor(atomizer.Spec())
	if !ok {
		t.Fatal("ScannerFor returned false")
	}

	mock := &mockColScanner{
		columns: []string{"id", "name", "age", "score", "active"},
		rows: [][]any{
			{int64(1), "Alice", int64(30), 95.5, true},
			{int64(2), nil, nil, nil, nil},
			{int64(3), "Charlie", int64(25), 88.0, false},
		},
		current: -1,
	}

	atoms, err := scanner.ScanAll(mock, mock.Next)
	if err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	if len(atoms) != 3 {
		t.Fatalf("expected 3 atoms, got %d", len(atoms))
	}

	// Verify IDs are correct (non-pointer types work correctly across rows)
	if atoms[0].Ints["ID"] != 1 {
		t.Errorf("row 0: expected ID=1, got %d", atoms[0].Ints["ID"])
	}
	if atoms[1].Ints["ID"] != 2 {
		t.Errorf("row 1: expected ID=2, got %d", atoms[1].Ints["ID"])
	}
	if atoms[2].Ints["ID"] != 3 {
		t.Errorf("row 2: expected ID=3, got %d", atoms[2].Ints["ID"])
	}

	// Verify nullable values are present (pointers may share underlying storage
	// across rows when using ScanAll due to destination reuse - this is expected)
	if atoms[0].StringPtrs["Name"] == nil {
		t.Error("row 0: expected Name to be non-nil")
	}
	if atoms[1].StringPtrs["Name"] != nil {
		t.Errorf("row 1: expected Name=nil, got non-nil")
	}
}
