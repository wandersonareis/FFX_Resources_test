package ddsphyre

import "fmt"

// namespace é o bloco auto-descritivo do PhyreEngine: classes, membros,
// descritores de tipo e a tabela de strings. Daqui saem TODOS os offsets —
// o layout da textura varia entre versões/formats e nada é hardcoded.
type namespace struct {
	buf          []byte
	stringTable  []byte
	typeDescOff  uint64
	classDescOff uint64
	memberOff    uint64
	typeCount    uint64
	classCount   uint64
	memberCount  uint64
}

// newNamespace decodifica o cabeçalho e localiza as seções internas
// (equivalente ao início de _getPhyreInfo).
func newNamespace(buf []byte) (*namespace, error) {
	if len(buf) < nsHeaderSize {
		return nil, fmt.Errorf("dds.phyre: namespace de %d bytes menor que o cabeçalho (%d)", len(buf), nsHeaderSize)
	}
	n := &namespace{buf: buf}
	var err error
	if n.typeCount, err = u64Of(buf, nsOffTypeCount); err != nil {
		return nil, err
	}
	if n.classCount, err = u64Of(buf, nsOffClassCount); err != nil {
		return nil, err
	}
	if n.memberCount, err = u64Of(buf, nsOffClassDataMemberCnt); err != nil {
		return nil, err
	}
	nsSize, err := u64Of(buf, nsOffSize)
	if err != nil {
		return nil, err
	}
	if nsSize > uint64(len(buf)) {
		return nil, fmt.Errorf("dds.phyre: namespace declara %d bytes, buffer tem %d", nsSize, len(buf))
	}
	stringTableSize, err := u64Of(buf, nsOffStringTableSize)
	if err != nil {
		return nil, err
	}
	defaultBufferCount, err := u64Of(buf, nsOffDefaultBufferCount)
	if err != nil {
		return nil, err
	}
	defaultBufferSize, err := u64Of(buf, nsOffDefaultBufferSize)
	if err != nil {
		return nil, err
	}

	// A tabela de strings fica no fim do namespace, antes dos buffers
	// padrão (mesmo cálculo do tool C++).
	stringTableStart := nsSize - stringTableSize - defaultBufferCount*defaultBufferSize
	if stringTableStart > nsSize || stringTableStart > uint64(len(buf)) {
		return nil, fmt.Errorf("dds.phyre: string table em offset %d inválido (namespace %d)", stringTableStart, nsSize)
	}
	n.stringTable = buf[int(stringTableStart):]

	n.typeDescOff = nsHeaderSize
	n.classDescOff = nsHeaderSize + 4*n.typeCount
	n.memberOff = n.classDescOff + classDescLen*n.classCount

	return n, nil
}

func u64Of(b []byte, off uint64) (uint64, error) {
	v, err := le32(b, off)
	if err != nil {
		return 0, err
	}
	return uint64(v), nil
}

// u32 lê um uint32 do namespace com checagem de faixa.
func (n *namespace) u32(off uint64) (uint32, error) {
	return le32(n.buf, off)
}

// cstr lê uma string da tabela de strings.
func (n *namespace) cstr(off uint64) (string, bool) {
	return cstr(n.stringTable, off)
}

// className resolve o nome da classe de um classId (1-based), como faz
// _getInstanceStartRelative no tool C++.
func (n *namespace) className(classID uint32) (string, error) {
	idx := uint64(classID) - 1
	if idx >= n.classCount {
		return "", fmt.Errorf("dds.phyre: classId %d fora do intervalo (classCount %d)", classID, n.classCount)
	}
	nameOff, err := n.u32(n.classDescOff + idx*classDescLen + classOffNameOffset)
	if err != nil {
		return "", fmt.Errorf("dds.phyre: lendo nome da classe %d: %w", classID, err)
	}
	name, ok := n.cstr(uint64(nameOff))
	if !ok {
		return "", fmt.Errorf("dds.phyre: nome da classe %d (offset %d) fora da string table", classID, nameOff)
	}
	return name, nil
}

// findClass localiza uma classe e devolve o offset do PRIMEIRO membro
// dela (a soma de dataMemberCount das classes anteriores) e a contagem.
// É o _findClass do tool C++.
func (n *namespace) findClass(className string) (memberStart, dataMemberCount uint64, ok bool) {
	var total uint64
	for i := uint64(0); i < n.classCount; i++ {
		nameOff, err := n.u32(n.classDescOff + i*classDescLen + classOffNameOffset)
		if err != nil {
			return 0, 0, false
		}
		if name, found := n.cstr(uint64(nameOff)); found && name == className {
			count, err := n.u32(n.classDescOff + i*classDescLen + classOffDataMemberCount)
			if err != nil {
				return 0, 0, false
			}
			return total, uint64(count), true
		}
		count, err := n.u32(n.classDescOff + i*classDescLen + classOffDataMemberCount)
		if err != nil {
			return 0, 0, false
		}
		total += uint64(count)
	}
	return 0, 0, false
}

// typeName resolve o nome de um type descriptor (0-based), como em
// `stringTable[typeDescriptors[typeId]]` do tool C++ — o descritor guarda
// um OFFSET na string table, não o nome.
func (n *namespace) typeName(typeID uint32) (string, error) {
	if uint64(typeID) >= n.typeCount {
		return "", fmt.Errorf("dds.phyre: typeId %d além de typeCount %d", typeID, n.typeCount)
	}
	off, err := n.u32(n.typeDescOff + uint64(typeID)*4)
	if err != nil {
		return "", fmt.Errorf("dds.phyre: lendo type descriptor %d: %w", typeID, err)
	}
	name, ok := n.cstr(uint64(off))
	if !ok {
		return "", fmt.Errorf("dds.phyre: type %d (offset %d) fora da string table", typeID, off)
	}
	return name, nil
}

// findMember devolve o valueOffset de um membro dentro do intervalo da
// classe (é o _findMember do tool C++).
func (n *namespace) findMember(memberStart, count uint64, memberName string) (uint64, error) {
	if memberStart+count > n.memberCount {
		return 0, fmt.Errorf("dds.phyre: intervalo de membros [%d,%d) além de %d", memberStart, memberStart+count, n.memberCount)
	}
	for i := uint64(0); i < count; i++ {
		base := n.memberOff + (memberStart+i)*memberDescLen
		nameOff, err := n.u32(base + memberOffNameOffset)
		if err != nil {
			return 0, fmt.Errorf("dds.phyre: lendo nome do membro: %w", err)
		}
		if name, ok := n.cstr(uint64(nameOff)); ok && name == memberName {
			valueOffset, err := n.u32(base + memberOffValueOffset)
			if err != nil {
				return 0, fmt.Errorf("dds.phyre: lendo valueOffset de %s: %w", memberName, err)
			}
			return uint64(valueOffset), nil
		}
	}
	return 0, fmt.Errorf("dds.phyre: membro %q não encontrado no namespace", memberName)
}
