// Copyright 2024 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package functionUtil

import (
	"github.com/jdkato/prose/v2"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"strings"
)

type Token struct {
	Value string
	Pos   int
}

func Tokenize(input string, format string, stopWords map[string]bool) ([]*Token, error) {

	switch strings.ToLower(format) {
	case "2gram":
		return nil, moerr.NewNYINoCtx("2gram not supported")
	case "3gram":
		return nil, moerr.NewNYINoCtx("3gram not supported")
	case "words":
		return tokenizeRegular(input, stopWords)
	default:
		return nil, moerr.NewNYINoCtx("tokenize format not supported")
	}
}

func tokenizeRegular(input string, stopWords map[string]bool) ([]*Token, error) {
	doc, err := prose.NewDocument(input)
	if err != nil {
		return nil, err
	}

	tokens := make([]*Token, len(doc.Tokens()))
	i := 0
	for pos, tok := range doc.Tokens() {
		if stopWords != nil {
			if _, ok := stopWords[tok.Text]; ok {
				continue
			}
		}

		tokens[i] = &Token{
			Value: tok.Text,
			Pos:   pos,
		}
		i++
	}

	return tokens, nil
}
