/*
Copyright 2014 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tile

import "testing"

func TestNewKey(t *testing.T) {
	cases := []struct {
		size, x, y int
	}{
		{0, 1<<14 - 1, 1<<14 - 1},
		{0, 1<<14 - 1, 0},
		{0, 0, 1<<14 - 1},
		{0, 0, 0},
		{1, 1, 1},
		{1, 2, 3},
		{2, 3, 1},
		{3, 3, 3},
		{5, 1<<14 - 1, 1<<14 - 1},
	}
	for i, tt := range cases {
		k := NewKey(byte(tt.size), uint16(tt.x), uint16(tt.y))
		if k.Size() != byte(tt.size) {
			t.Errorf("%d. size = %d; want %d", i, k.Size(), tt.size)
		}
		if k.X() != uint16(tt.x) {
			t.Errorf("%d. x = %d; want %d", i, k.X(), tt.x)
		}
		if k.Y() != uint16(tt.y) {
			t.Errorf("%d. y = %d; want %d", i, k.Y(), tt.y)
		}
	}
}

// TestKeysSortBySizeThenPosition guards the ordering the lookup's binary search
// depends on.
func TestKeysSortBySizeThenPosition(t *testing.T) {
	if a, b := NewKey(0, 5, 5), NewKey(1, 0, 0); a >= b {
		t.Errorf("NewKey(0,5,5) = %#x should sort before NewKey(1,0,0) = %#x", a, b)
	}
	if a, b := NewKey(2, 0, 1), NewKey(2, 0, 2); a >= b {
		t.Errorf("NewKey(2,0,1) = %#x should sort before NewKey(2,0,2) = %#x", a, b)
	}
	if a, b := NewKey(2, 1, 7), NewKey(2, 2, 7); a >= b {
		t.Errorf("NewKey(2,1,7) = %#x should sort before NewKey(2,2,7) = %#x", a, b)
	}
}

func TestOceanIndexCannotCollide(t *testing.T) {
	if OceanIndex != 1<<16-1 {
		t.Errorf("OceanIndex = %#x; want the largest uint16 so a real leaf index can never reach it", OceanIndex)
	}
}
