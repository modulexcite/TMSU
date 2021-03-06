// Copyright 2011-2015 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package storage

import (
	"errors"
	"fmt"
	"tmsu/entities"
	"tmsu/storage/database"
	"unicode"
)

// The number of tags in the database.
func (storage *Storage) TagCount(tx *Tx) (uint, error) {
	return database.TagCount(tx.tx)
}

// The set of tags.
func (storage *Storage) Tags(tx *Tx) (entities.Tags, error) {
	return database.Tags(tx.tx)
}

// Retrieves a specific tag.
func (storage Storage) Tag(tx *Tx, id entities.TagId) (*entities.Tag, error) {
	return database.Tag(tx.tx, id)
}

// Retrieves a specific set of tags.
func (storage Storage) TagsByIds(tx *Tx, ids entities.TagIds) (entities.Tags, error) {
	return database.TagsByIds(tx.tx, ids)
}

// Retrieves a specific tag.
func (storage Storage) TagByName(tx *Tx, name string) (*entities.Tag, error) {
	return database.TagByName(tx.tx, name)
}

// Retrieves the set of named tags.
func (storage Storage) TagsByNames(tx *Tx, names []string) (entities.Tags, error) {
	return database.TagsByNames(tx.tx, names)
}

// Adds a tag.
func (storage *Storage) AddTag(tx *Tx, name string) (*entities.Tag, error) {
	if err := validateTagName(name); err != nil {
		return nil, err
	}

	return database.InsertTag(tx.tx, name)
}

// Renames a tag.
func (storage Storage) RenameTag(tx *Tx, tagId entities.TagId, name string) (*entities.Tag, error) {
	if err := validateTagName(name); err != nil {
		return nil, err
	}

	return database.RenameTag(tx.tx, tagId, name)
}

// Copies a tag.
func (storage Storage) CopyTag(tx *Tx, sourceTagId entities.TagId, name string) (*entities.Tag, error) {
	if err := validateTagName(name); err != nil {
		return nil, err
	}

	tag, err := database.InsertTag(tx.tx, name)
	if err != nil {
		return nil, fmt.Errorf("could not create tag '%v': %v", name, err)
	}

	err = database.CopyFileTags(tx.tx, sourceTagId, tag.Id)
	if err != nil {
		return nil, fmt.Errorf("could not copy file tags for tag #%v to tag '%v': %v", sourceTagId, name, err)
	}

	return tag, nil
}

// Deletes a tag.
func (storage Storage) DeleteTag(tx *Tx, tagId entities.TagId) error {
	err := storage.DeleteFileTagsByTagId(tx, tagId)
	if err != nil {
		return err
	}

	err = database.DeleteTag(tx.tx, tagId)
	if err != nil {
		return fmt.Errorf("could not delete tag '%v': %v", tagId, err)
	}

	return nil
}

// Retrieves the tag usage.
func (storage Storage) TagUsage(tx *Tx) ([]entities.TagFileCount, error) {
	return database.TagUsage(tx.tx)
}

// unexported

var validTagChars = []*unicode.RangeTable{unicode.Letter, unicode.Number, unicode.Punct, unicode.Symbol}

func validateTagName(tagName string) error {
	switch tagName {
	case "":
		return errors.New("tag name cannot be empty")
	case ".", "..":
		return errors.New("tag name cannot be '.' or '..'") // cannot be used in the VFS
	case "and", "AND", "or", "OR", "not", "NOT":
		return errors.New("tag name cannot be a logical operator: 'and', 'or' or 'not'") // used in query language
	case "eq", "EQ", "ne", "NE", "lt", "LT", "gt", "GT", "le", "LE", "ge", "GE":
		return errors.New("tag name cannot be a comparison operator: 'eq', 'ne', 'gt', 'lt', 'ge' or 'le'") // used in query language
	}

	if tagName[0] == '-' {
		return errors.New("tag name cannot start with a minus: '-'") // used in query language
	}

	for _, ch := range tagName {
		switch ch {
		case '(', ')':
			return errors.New("tag names cannot contain parentheses: '(' or ')'") // used in query language
		case ',':
			return errors.New("tag names cannot contain comma: ','") // reserved for tag delimiter
		case '=', '!', '<', '>':
			return errors.New("tag names cannot contain a comparison operator: '=', '!', '<' or '>'") // reserved for tag values
		case ' ', '\t':
			return errors.New("tag names cannot contain space or tab") // used as tag delimiter
		case '/':
			return errors.New("tag names cannot contain slash: '/'") // cannot be used in the VFS
		}

		if !unicode.IsOneOf(validTagChars, ch) {
			return fmt.Errorf("tag names cannot contain '%c'", ch)
		}
	}

	return nil
}
