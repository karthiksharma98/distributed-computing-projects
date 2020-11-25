package main

func (m *Mapler) Maple(input string) error {
	m.Emit(input, "1")
	return nil
}
