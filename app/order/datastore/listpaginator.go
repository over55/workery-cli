package datastore

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/bartmika/timekit"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	SortOrderAscending  = 1
	SortOrderDescending = -1
)

type OrderPaginationListFilter struct {
	// Pagination related.
	Cursor    string
	PageSize  int64
	SortField string
	SortOrder int8 // 1=ascending | -1=descending

	// Filter related.
	TenantID         primitive.ObjectID
	CustomerID       primitive.ObjectID
	AssociateID      primitive.ObjectID
	Status           int8
	Type             int8
	ExcludeArchived  bool
	SearchText       string
	ModifiedByUserID primitive.ObjectID
}

// OrderPaginationLiteListResult represents the paginated list results for
// the order lite records (meaning limited).
type OrderPaginationLiteListResult struct {
	Results     []*OrderLite `json:"results"`
	NextCursor  string       `json:"next_cursor"`
	HasNextPage bool         `json:"has_next_page"`
}

// newPaginationFilter will create the mongodb filter to apply the cursor or
// or ignore it depending if a cursor was specified in the filter.
func (impl OrderStorerImpl) newPaginationFilter(f *OrderPaginationListFilter) (bson.M, error) {
	if len(f.Cursor) > 0 {
		// STEP 1: Decode the cursor which is encoded in a base64 format.
		decodedCursor, err := base64.RawStdEncoding.DecodeString(f.Cursor)
		if err != nil {
			return bson.M{}, fmt.Errorf("Failed to decode string: %v", err)
		}

		// STEP 2: Pick the specific cursor to build or else error.
		switch f.SortField {
		case "customer_lexical_name", "associate_lexical_name":
			// STEP 3: Build specific field for string.
			return impl.newPaginationFilterBasedOnString(f, string(decodedCursor))
		case "assignment_date", "start_date":
			// STEP 3: Build specific field of time.
			return impl.newPaginationFilterBasedOnTimestamp(f, string(decodedCursor))
		default:
			return nil, fmt.Errorf("unsupported sort field for `%v`, only supported fields are `associate_lexical_name`, `customer_lexical_name`, `start_date` and `assignment_date`", f.SortField)
		}
	}
	return bson.M{}, nil
}

func (impl OrderStorerImpl) newPaginationFilterBasedOnString(f *OrderPaginationListFilter, decodedCursor string) (bson.M, error) {
	// Extract our cursor into two parts which we need to use.
	arr := strings.Split(decodedCursor, "|")
	if len(arr) < 1 {
		return nil, fmt.Errorf("cursor is corrupted for the value `%v`", decodedCursor)
	}

	// The first part will contain the name we left off at. The second part will
	// be last ID we left off at.
	str := arr[0]
	lastID, err := primitive.ObjectIDFromHex(arr[1])
	if err != nil {
		return bson.M{}, fmt.Errorf("Failed to convert into mongodb object id: %v, from the decoded cursor of: %v", err, decodedCursor)
	}

	switch f.SortOrder {
	case SortOrderAscending:
		filter := bson.M{}
		filter["$or"] = []bson.M{
			bson.M{f.SortField: bson.M{"$gt": str}},
			bson.M{f.SortField: str, "_id": bson.M{"$gt": lastID}},
		}
		return filter, nil
	case SortOrderDescending:
		filter := bson.M{}
		filter["$or"] = []bson.M{
			bson.M{f.SortField: bson.M{"$lt": str}},
			bson.M{f.SortField: str, "_id": bson.M{"$lt": lastID}},
		}
		return filter, nil
	default:
		return nil, fmt.Errorf("unsupported sort order for `%v`, only supported values are `1` or `-1`", f.SortOrder)
	}
}

func (impl OrderStorerImpl) newPaginationFilterBasedOnTimestamp(f *OrderPaginationListFilter, decodedCursor string) (bson.M, error) {
	// Extract our cursor into two parts which we need to use.
	arr := strings.Split(decodedCursor, "|")
	if len(arr) < 1 {
		return nil, fmt.Errorf("cursor is corrupted for the value `%v`", decodedCursor)
	}

	// The first part will contain the name we left off at. The second part will
	// be last ID we left off at.
	timestampStr := arr[0]
	lastID, err := primitive.ObjectIDFromHex(arr[1])
	if err != nil {
		return nil, fmt.Errorf("Failed to convert into mongodb object id: %v, from the decoded cursor of: %v", err, decodedCursor)
	}

	timestamp, err := timekit.ParseJavaScriptTimeString(timestampStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse javascript timestamp: `%v`", err)
	}

	switch f.SortOrder {
	case SortOrderAscending:
		filter := bson.M{}
		filter["$or"] = []bson.M{
			bson.M{f.SortField: bson.M{"$gt": timestamp}},
			bson.M{f.SortField: timestamp, "_id": bson.M{"$gt": lastID}},
		}
		return filter, nil
	case SortOrderDescending:
		filter := bson.M{}
		filter["$or"] = []bson.M{
			bson.M{f.SortField: bson.M{"$lt": timestamp}},
			bson.M{f.SortField: timestamp, "_id": bson.M{"$lt": lastID}},
		}
		return filter, nil
	default:
		return nil, fmt.Errorf("unsupported sort order for `%v`, only supported values are `1` or `-1`", f.SortOrder)
	}
}

// newPaginatorOptions will generate the mongodb options which will support the
// paginator in ordering the data to work.
func (impl OrderStorerImpl) newPaginationOptions(f *OrderPaginationListFilter) (*options.FindOptions, error) {
	options := options.Find().
		SetSort(bson.D{
			{f.SortField, f.SortOrder},
			{"_id", f.SortOrder}, // Include _id in sorting for consistency
		}).
		SetLimit(f.PageSize)
	return options, nil
}

// newPaginatorNextCursor will return the base64 encoded next cursor which works
// with our paginator.
func (impl OrderStorerImpl) newPaginatorNextCursor(f *OrderPaginationListFilter, results []*OrderLite) (string, error) {
	var lastDatum *OrderLite

	// Remove the extra document from the current page
	results = results[:len(results)]

	// Get the last document's _id as the next cursor
	lastDatum = results[len(results)-1]

	// Variable used to store the next cursor.
	var nextCursor string

	switch f.SortField {
	case "customer_lexical_name":
		nextCursor = fmt.Sprintf("%v|%v", lastDatum.CustomerLexicalName, lastDatum.ID.Hex())
		break
	case "associate_lexical_name":
		// Generate the unique next cursor.
		nextCursor = fmt.Sprintf("%v|%v", lastDatum.AssociateLexicalName, lastDatum.ID.Hex())
		break
	case "assignment_date":
		timestamp := lastDatum.AssignmentDate.UnixMilli()
		nextCursor = fmt.Sprintf("%v|%v", timestamp, lastDatum.ID.Hex())
		break
	case "start_date":
		timestamp := lastDatum.StartDate.UnixMilli()
		nextCursor = fmt.Sprintf("%v|%v", timestamp, lastDatum.ID.Hex())
		break
	default:
		return "", fmt.Errorf("unsupported sort field in options for `%v`, only supported fields are `customer_lexical_name` and `assignment_date`", f.SortField)
	}

	// Encode to base64 without the `=` symbol that would corrupt when we
	// use the http url argument. Special thanks to:
	// https://www.golinuxcloud.com/golang-base64-encode/
	encoded := base64.RawStdEncoding.EncodeToString([]byte(nextCursor))

	return encoded, nil
}
