package tui

import (
	"github.com/amureki/metabase-explorer/pkg/api"
)

// Messages for Bubble Tea updates

type databasesLoaded struct {
	databases []api.Database
	err       error
}

type schemasLoaded struct {
	schemas []api.Schema
	err     error
}

type tablesLoaded struct {
	tables []api.Table
	err    error
}

type fieldsLoaded struct {
	fields []api.Field
	err    error
}

type versionChecked struct {
	latestVersion string
	err           error
}

type spinnerTick struct{}

type connectionTested struct {
	err error
}

type collectionsLoaded struct {
	collections []api.Collection
	err         error
}

type collectionItemsLoaded struct {
	items []api.CollectionItem
	err   error
}

type cardDetailLoaded struct {
	detail *api.CardDetail
	err    error
}
