package main

import "testing"

func TestExpandTextMacro(t *testing.T) {
    cases:=[]struct{in,want string}{
        {"안녕하세요 😀","안녕하세요 😀"},
        {`첫줄\n둘째줄`,"첫줄\n둘째줄"},
        {`문장\t다음`,"문장\t다음"},
        {`C:\Users\Public`,`C:\Users\Public`},
        {`값\\n그대로`,`값\n그대로`},
        {"한글\r\n다음","한글\n다음"},
    }
    for _,tc:=range cases {
        got,err:=expandTextMacro(tc.in)
        if err!=nil||got!=tc.want {t.Errorf("expand %q => %q (%v), want %q",tc.in,got,err,tc.want)}
    }
}
func TestExpandTextMacroRejectsUnsafeInput(t *testing.T) {
    for _,s:=range []string{"",string([]byte{0xff}),"a\x00b",string(make([]byte,4097))} {
        if _,err:=expandTextMacro(s);err==nil {t.Errorf("expected rejection for %q",s)}
    }
}

func TestMultilineEditorNewlines(t *testing.T) {
    // Windows EDIT exposes CRLF; pasting may also introduce lone CR.
    saved := normalizeEditorNewlines("첫 줄\r\n둘째 줄\r셋째 줄\n마지막 😀")
    want := "첫 줄\n둘째 줄\n셋째 줄\n마지막 😀"
    if saved != want {
        t.Fatalf("line ending normalization: got %q, want %q", saved, want)
    }
    expanded, err := expandTextMacro(saved)
    if err != nil || expanded != want {
        t.Fatalf("multiline macro: got %q (%v), want %q", expanded, err, want)
    }
}
