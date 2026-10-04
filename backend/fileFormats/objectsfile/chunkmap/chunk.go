// Package chunkmap implementa mapeamento struct <-> binário no estilo das
// structs [StructLayout] do Fahrenheit (C#):
// https://github.com/fahrenheit-crew/fahrenheit (src/core/ffx*, src/core/ffx2)
//
// Cada arquivo binário é modelado por UMA struct T que cobre todos os bytes do
// chunk. Campos de texto são TextRef/TextPair (models.Segment) e são os ÚNICOS
// que o fluxo de texto exporta e importa; todo o restante é apenas preservado.
// Sobras/padding que a struct não nomeia ficam em Chunk.Tail e são devolvidas
// byte-a-byte no Encode (round-trip fiel).
package chunkmap

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"slices"
)

// Chunk é um chunk binário decodificado: a struct Value cobre os primeiros
// binary.Size(Value) bytes e Tail guarda o restante (clonado — o leitor V2
// reutiliza o buffer entre objetos).
type Chunk[T any] struct {
	Value T
	Tail  []byte
}

// Decode lê data em T (little-endian). Os bytes além da struct vão para Tail.
// Erros: tipo não serializável, chunk menor que a struct, falha de leitura.
func Decode[T any](data []byte) (*Chunk[T], error) {
	var v T
	size := binary.Size(v)
	if size < 0 {
		return nil, fmt.Errorf("chunkmap: tipo %T nao suportado por encoding/binary", v)
	}
	if len(data) < size {
		return nil, fmt.Errorf("chunkmap: chunk de %d bytes menor que a struct (%d)", len(data), size)
	}
	if err := binary.Read(bytes.NewReader(data[:size]), binary.LittleEndian, &v); err != nil {
		return nil, fmt.Errorf("chunkmap: decodificar %T: %w", v, err)
	}
	return &Chunk[T]{Value: v, Tail: slices.Clone(data[size:])}, nil
}

// Encode serializa Value de volta (little-endian) e anexa Tail.
// Garante binary.Size(Value)+len(Tail) bytes — round-trip fiel de Decode.
func (c *Chunk[T]) Encode() ([]byte, error) {
	size := binary.Size(c.Value)
	if size < 0 {
		return nil, fmt.Errorf("chunkmap: tipo %T nao suportado por encoding/binary", c.Value)
	}
	buf := bytes.NewBuffer(make([]byte, 0, size+len(c.Tail)))
	if err := binary.Write(buf, binary.LittleEndian, &c.Value); err != nil {
		return nil, fmt.Errorf("chunkmap: codificar %T: %w", c.Value, err)
	}
	buf.Write(c.Tail)
	return buf.Bytes(), nil
}

// StructSize devolve o tamanho binário de T (para validação contra o
// individualLength do cabeçalho do arquivo real).
func StructSize[T any]() int {
	return binary.Size(*new(T))
}
