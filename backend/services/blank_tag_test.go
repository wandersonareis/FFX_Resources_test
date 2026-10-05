package services

import "testing"

func TestIsBlankText_NewlineTag(t *testing.T) {
	if !isBlankText("{TEXT_NEWLINE}") {
		t.Fatal("novo: {TEXT_NEWLINE} deve virar em branco")
	}
	if !isBlankText(" {TEXT_NEWLINE} ") {
		t.Fatal("novo: vazio+newline deve ser branco")
	}
	if isBlankText("Data Transfer") {
		t.Fatal("Data Transfer não é branco")
	}
	if isBlankText("{TEXT_NEWLINE}Data Transfer") {
		t.Fatal("Data Transfer com newline não é branco")
	}
	if !isBlankText("{TEXT_ITALIC}{TEXT_NEWLINE}{TEXT_NORMAL}") {
		t.Fatal("só tags de forma somem")
	}
	if isBlankText("{ICON:01}") {
		t.Fatal("tag de controle não é blank")
	}
}
