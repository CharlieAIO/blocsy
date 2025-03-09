// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package types

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson89aae3efDecodeBlocsyInternalTypes(in *jlexer.Lexer, out *TrackerResponse) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "data":
			easyjson89aae3efDecode(in, &out.Data)
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson89aae3efEncodeBlocsyInternalTypes(out *jwriter.Writer, in TrackerResponse) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"data\":"
		out.RawString(prefix[1:])
		easyjson89aae3efEncode(out, in.Data)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v TrackerResponse) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson89aae3efEncodeBlocsyInternalTypes(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v TrackerResponse) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson89aae3efEncodeBlocsyInternalTypes(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *TrackerResponse) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson89aae3efDecodeBlocsyInternalTypes(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *TrackerResponse) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson89aae3efDecodeBlocsyInternalTypes(l, v)
}
func easyjson89aae3efDecode(in *jlexer.Lexer, out *struct {
	Amount   string `json:"amount"`
	Base     string `json:"base"`
	Currency string `json:"currency"`
}) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "amount":
			out.Amount = string(in.String())
		case "base":
			out.Base = string(in.String())
		case "currency":
			out.Currency = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson89aae3efEncode(out *jwriter.Writer, in struct {
	Amount   string `json:"amount"`
	Base     string `json:"base"`
	Currency string `json:"currency"`
}) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"amount\":"
		out.RawString(prefix[1:])
		out.String(string(in.Amount))
	}
	{
		const prefix string = ",\"base\":"
		out.RawString(prefix)
		out.String(string(in.Base))
	}
	{
		const prefix string = ",\"currency\":"
		out.RawString(prefix)
		out.String(string(in.Currency))
	}
	out.RawByte('}')
}
func easyjson89aae3efDecodeBlocsyInternalTypes1(in *jlexer.Lexer, out *TokenAccountDetails) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "UserAccount":
			out.UserAccount = string(in.String())
		case "MintAddress":
			out.MintAddress = string(in.String())
		case "Decimals":
			out.Decimals = int(in.Int())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson89aae3efEncodeBlocsyInternalTypes1(out *jwriter.Writer, in TokenAccountDetails) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"UserAccount\":"
		out.RawString(prefix[1:])
		out.String(string(in.UserAccount))
	}
	{
		const prefix string = ",\"MintAddress\":"
		out.RawString(prefix)
		out.String(string(in.MintAddress))
	}
	{
		const prefix string = ",\"Decimals\":"
		out.RawString(prefix)
		out.Int(int(in.Decimals))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v TokenAccountDetails) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson89aae3efEncodeBlocsyInternalTypes1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v TokenAccountDetails) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson89aae3efEncodeBlocsyInternalTypes1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *TokenAccountDetails) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson89aae3efDecodeBlocsyInternalTypes1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *TokenAccountDetails) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson89aae3efDecodeBlocsyInternalTypes1(l, v)
}
func easyjson89aae3efDecodeBlocsyInternalTypes2(in *jlexer.Lexer, out *Token) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "name":
			out.Name = string(in.String())
		case "symbol":
			out.Symbol = string(in.String())
		case "decimals":
			out.Decimals = uint8(in.Uint8())
		case "address":
			out.Address = string(in.String())
		case "supply":
			out.Supply = string(in.String())
		case "createdBlock":
			out.CreatedBlock = int64(in.Int64())
		case "network":
			out.Network = string(in.String())
		case "createdTimestamp":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.CreatedTimestamp).UnmarshalJSON(data))
			}
		case "deployer":
			if in.IsNull() {
				in.Skip()
				out.Deployer = nil
			} else {
				if out.Deployer == nil {
					out.Deployer = new(string)
				}
				*out.Deployer = string(in.String())
			}
		case "metadata":
			if in.IsNull() {
				in.Skip()
				out.Metadata = nil
			} else {
				if out.Metadata == nil {
					out.Metadata = new(string)
				}
				*out.Metadata = string(in.String())
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson89aae3efEncodeBlocsyInternalTypes2(out *jwriter.Writer, in Token) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix[1:])
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"symbol\":"
		out.RawString(prefix)
		out.String(string(in.Symbol))
	}
	{
		const prefix string = ",\"decimals\":"
		out.RawString(prefix)
		out.Uint8(uint8(in.Decimals))
	}
	{
		const prefix string = ",\"address\":"
		out.RawString(prefix)
		out.String(string(in.Address))
	}
	{
		const prefix string = ",\"supply\":"
		out.RawString(prefix)
		out.String(string(in.Supply))
	}
	{
		const prefix string = ",\"createdBlock\":"
		out.RawString(prefix)
		out.Int64(int64(in.CreatedBlock))
	}
	{
		const prefix string = ",\"network\":"
		out.RawString(prefix)
		out.String(string(in.Network))
	}
	{
		const prefix string = ",\"createdTimestamp\":"
		out.RawString(prefix)
		out.Raw((in.CreatedTimestamp).MarshalJSON())
	}
	if in.Deployer != nil {
		const prefix string = ",\"deployer\":"
		out.RawString(prefix)
		out.String(string(*in.Deployer))
	}
	if in.Metadata != nil {
		const prefix string = ",\"metadata\":"
		out.RawString(prefix)
		out.String(string(*in.Metadata))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Token) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson89aae3efEncodeBlocsyInternalTypes2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Token) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson89aae3efEncodeBlocsyInternalTypes2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Token) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson89aae3efDecodeBlocsyInternalTypes2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Token) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson89aae3efDecodeBlocsyInternalTypes2(l, v)
}
func easyjson89aae3efDecodeBlocsyInternalTypes3(in *jlexer.Lexer, out *QuoteTokenSimple) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "identifier":
			out.Identifier = string(in.String())
		case "address":
			out.Address = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson89aae3efEncodeBlocsyInternalTypes3(out *jwriter.Writer, in QuoteTokenSimple) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"identifier\":"
		out.RawString(prefix[1:])
		out.String(string(in.Identifier))
	}
	{
		const prefix string = ",\"address\":"
		out.RawString(prefix)
		out.String(string(in.Address))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v QuoteTokenSimple) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson89aae3efEncodeBlocsyInternalTypes3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v QuoteTokenSimple) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson89aae3efEncodeBlocsyInternalTypes3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *QuoteTokenSimple) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson89aae3efDecodeBlocsyInternalTypes3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *QuoteTokenSimple) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson89aae3efDecodeBlocsyInternalTypes3(l, v)
}
func easyjson89aae3efDecodeBlocsyInternalTypes4(in *jlexer.Lexer, out *QuoteToken) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "identifier":
			out.Identifier = string(in.String())
		case "name":
			out.Name = string(in.String())
		case "symbol":
			out.Symbol = string(in.String())
		case "address":
			out.Address = string(in.String())
		case "decimals":
			out.Decimals = uint8(in.Uint8())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson89aae3efEncodeBlocsyInternalTypes4(out *jwriter.Writer, in QuoteToken) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"identifier\":"
		out.RawString(prefix[1:])
		out.String(string(in.Identifier))
	}
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"symbol\":"
		out.RawString(prefix)
		out.String(string(in.Symbol))
	}
	{
		const prefix string = ",\"address\":"
		out.RawString(prefix)
		out.String(string(in.Address))
	}
	{
		const prefix string = ",\"decimals\":"
		out.RawString(prefix)
		out.Uint8(uint8(in.Decimals))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v QuoteToken) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson89aae3efEncodeBlocsyInternalTypes4(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v QuoteToken) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson89aae3efEncodeBlocsyInternalTypes4(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *QuoteToken) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson89aae3efDecodeBlocsyInternalTypes4(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *QuoteToken) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson89aae3efDecodeBlocsyInternalTypes4(l, v)
}
func easyjson89aae3efDecodeBlocsyInternalTypes5(in *jlexer.Lexer, out *QueryAll) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "source":
			out.Source = string(in.String())
		case "wallet":
			if in.IsNull() {
				in.Skip()
				out.Wallet = nil
			} else {
				if out.Wallet == nil {
					out.Wallet = new(string)
				}
				*out.Wallet = string(in.String())
			}
		case "token":
			if in.IsNull() {
				in.Skip()
				out.Token = nil
			} else {
				if out.Token == nil {
					out.Token = new(string)
				}
				*out.Token = string(in.String())
			}
		case "name":
			if in.IsNull() {
				in.Skip()
				out.Name = nil
			} else {
				if out.Name == nil {
					out.Name = new(string)
				}
				*out.Name = string(in.String())
			}
		case "symbol":
			if in.IsNull() {
				in.Skip()
				out.Symbol = nil
			} else {
				if out.Symbol == nil {
					out.Symbol = new(string)
				}
				*out.Symbol = string(in.String())
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson89aae3efEncodeBlocsyInternalTypes5(out *jwriter.Writer, in QueryAll) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"source\":"
		out.RawString(prefix[1:])
		out.String(string(in.Source))
	}
	if in.Wallet != nil {
		const prefix string = ",\"wallet\":"
		out.RawString(prefix)
		out.String(string(*in.Wallet))
	}
	if in.Token != nil {
		const prefix string = ",\"token\":"
		out.RawString(prefix)
		out.String(string(*in.Token))
	}
	if in.Name != nil {
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(*in.Name))
	}
	if in.Symbol != nil {
		const prefix string = ",\"symbol\":"
		out.RawString(prefix)
		out.String(string(*in.Symbol))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v QueryAll) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson89aae3efEncodeBlocsyInternalTypes5(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v QueryAll) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson89aae3efEncodeBlocsyInternalTypes5(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *QueryAll) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson89aae3efDecodeBlocsyInternalTypes5(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *QueryAll) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson89aae3efDecodeBlocsyInternalTypes5(l, v)
}
func easyjson89aae3efDecodeBlocsyInternalTypes6(in *jlexer.Lexer, out *ProcessInstructionData) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "AccountKeys":
			if in.IsNull() {
				in.Skip()
				out.AccountKeys = nil
			} else {
				in.Delim('[')
				if out.AccountKeys == nil {
					if !in.IsDelim(']') {
						out.AccountKeys = make([]string, 0, 4)
					} else {
						out.AccountKeys = []string{}
					}
				} else {
					out.AccountKeys = (out.AccountKeys)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.AccountKeys = append(out.AccountKeys, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "InstructionAccounts":
			if in.IsNull() {
				in.Skip()
				out.InstructionAccounts = nil
			} else {
				if out.InstructionAccounts == nil {
					out.InstructionAccounts = new([]int)
				}
				if in.IsNull() {
					in.Skip()
					*out.InstructionAccounts = nil
				} else {
					in.Delim('[')
					if *out.InstructionAccounts == nil {
						if !in.IsDelim(']') {
							*out.InstructionAccounts = make([]int, 0, 8)
						} else {
							*out.InstructionAccounts = []int{}
						}
					} else {
						*out.InstructionAccounts = (*out.InstructionAccounts)[:0]
					}
					for !in.IsDelim(']') {
						var v2 int
						v2 = int(in.Int())
						*out.InstructionAccounts = append(*out.InstructionAccounts, v2)
						in.WantComma()
					}
					in.Delim(']')
				}
			}
		case "Accounts":
			if in.IsNull() {
				in.Skip()
				out.Accounts = nil
			} else {
				if out.Accounts == nil {
					out.Accounts = new([]int)
				}
				if in.IsNull() {
					in.Skip()
					*out.Accounts = nil
				} else {
					in.Delim('[')
					if *out.Accounts == nil {
						if !in.IsDelim(']') {
							*out.Accounts = make([]int, 0, 8)
						} else {
							*out.Accounts = []int{}
						}
					} else {
						*out.Accounts = (*out.Accounts)[:0]
					}
					for !in.IsDelim(']') {
						var v3 int
						v3 = int(in.Int())
						*out.Accounts = append(*out.Accounts, v3)
						in.WantComma()
					}
					in.Delim(']')
				}
			}
		case "Transfers":
			if in.IsNull() {
				in.Skip()
				out.Transfers = nil
			} else {
				in.Delim('[')
				if out.Transfers == nil {
					if !in.IsDelim(']') {
						out.Transfers = make([]SolTransfer, 0, 0)
					} else {
						out.Transfers = []SolTransfer{}
					}
				} else {
					out.Transfers = (out.Transfers)[:0]
				}
				for !in.IsDelim(']') {
					var v4 SolTransfer
					(v4).UnmarshalEasyJSON(in)
					out.Transfers = append(out.Transfers, v4)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "ProgramId":
			if in.IsNull() {
				in.Skip()
				out.ProgramId = nil
			} else {
				if out.ProgramId == nil {
					out.ProgramId = new(string)
				}
				*out.ProgramId = string(in.String())
			}
		case "InnerProgramId":
			if in.IsNull() {
				in.Skip()
				out.InnerProgramId = nil
			} else {
				if out.InnerProgramId == nil {
					out.InnerProgramId = new(string)
				}
				*out.InnerProgramId = string(in.String())
			}
		case "InnerInstructionIndex":
			out.InnerInstructionIndex = int(in.Int())
		case "InnerIndex":
			if in.IsNull() {
				in.Skip()
				out.InnerIndex = nil
			} else {
				if out.InnerIndex == nil {
					out.InnerIndex = new(int)
				}
				*out.InnerIndex = int(in.Int())
			}
		case "Data":
			if in.IsNull() {
				in.Skip()
				out.Data = nil
			} else {
				if out.Data == nil {
					out.Data = new(string)
				}
				*out.Data = string(in.String())
			}
		case "InnerAccounts":
			if in.IsNull() {
				in.Skip()
				out.InnerAccounts = nil
			} else {
				if out.InnerAccounts == nil {
					out.InnerAccounts = new([]int)
				}
				if in.IsNull() {
					in.Skip()
					*out.InnerAccounts = nil
				} else {
					in.Delim('[')
					if *out.InnerAccounts == nil {
						if !in.IsDelim(']') {
							*out.InnerAccounts = make([]int, 0, 8)
						} else {
							*out.InnerAccounts = []int{}
						}
					} else {
						*out.InnerAccounts = (*out.InnerAccounts)[:0]
					}
					for !in.IsDelim(']') {
						var v5 int
						v5 = int(in.Int())
						*out.InnerAccounts = append(*out.InnerAccounts, v5)
						in.WantComma()
					}
					in.Delim(']')
				}
			}
		case "Logs":
			if in.IsNull() {
				in.Skip()
				out.Logs = nil
			} else {
				in.Delim('[')
				if out.Logs == nil {
					if !in.IsDelim(']') {
						out.Logs = make([]string, 0, 4)
					} else {
						out.Logs = []string{}
					}
				} else {
					out.Logs = (out.Logs)[:0]
				}
				for !in.IsDelim(']') {
					var v6 string
					v6 = string(in.String())
					out.Logs = append(out.Logs, v6)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "TokenAccountMap":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				out.TokenAccountMap = make(map[string]TokenAccountDetails)
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v7 TokenAccountDetails
					(v7).UnmarshalEasyJSON(in)
					(out.TokenAccountMap)[key] = v7
					in.WantComma()
				}
				in.Delim('}')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson89aae3efEncodeBlocsyInternalTypes6(out *jwriter.Writer, in ProcessInstructionData) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"AccountKeys\":"
		out.RawString(prefix[1:])
		if in.AccountKeys == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v8, v9 := range in.AccountKeys {
				if v8 > 0 {
					out.RawByte(',')
				}
				out.String(string(v9))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"InstructionAccounts\":"
		out.RawString(prefix)
		if in.InstructionAccounts == nil {
			out.RawString("null")
		} else {
			if *in.InstructionAccounts == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
				out.RawString("null")
			} else {
				out.RawByte('[')
				for v10, v11 := range *in.InstructionAccounts {
					if v10 > 0 {
						out.RawByte(',')
					}
					out.Int(int(v11))
				}
				out.RawByte(']')
			}
		}
	}
	{
		const prefix string = ",\"Accounts\":"
		out.RawString(prefix)
		if in.Accounts == nil {
			out.RawString("null")
		} else {
			if *in.Accounts == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
				out.RawString("null")
			} else {
				out.RawByte('[')
				for v12, v13 := range *in.Accounts {
					if v12 > 0 {
						out.RawByte(',')
					}
					out.Int(int(v13))
				}
				out.RawByte(']')
			}
		}
	}
	{
		const prefix string = ",\"Transfers\":"
		out.RawString(prefix)
		if in.Transfers == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v14, v15 := range in.Transfers {
				if v14 > 0 {
					out.RawByte(',')
				}
				(v15).MarshalEasyJSON(out)
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"ProgramId\":"
		out.RawString(prefix)
		if in.ProgramId == nil {
			out.RawString("null")
		} else {
			out.String(string(*in.ProgramId))
		}
	}
	{
		const prefix string = ",\"InnerProgramId\":"
		out.RawString(prefix)
		if in.InnerProgramId == nil {
			out.RawString("null")
		} else {
			out.String(string(*in.InnerProgramId))
		}
	}
	{
		const prefix string = ",\"InnerInstructionIndex\":"
		out.RawString(prefix)
		out.Int(int(in.InnerInstructionIndex))
	}
	{
		const prefix string = ",\"InnerIndex\":"
		out.RawString(prefix)
		if in.InnerIndex == nil {
			out.RawString("null")
		} else {
			out.Int(int(*in.InnerIndex))
		}
	}
	{
		const prefix string = ",\"Data\":"
		out.RawString(prefix)
		if in.Data == nil {
			out.RawString("null")
		} else {
			out.String(string(*in.Data))
		}
	}
	{
		const prefix string = ",\"InnerAccounts\":"
		out.RawString(prefix)
		if in.InnerAccounts == nil {
			out.RawString("null")
		} else {
			if *in.InnerAccounts == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
				out.RawString("null")
			} else {
				out.RawByte('[')
				for v16, v17 := range *in.InnerAccounts {
					if v16 > 0 {
						out.RawByte(',')
					}
					out.Int(int(v17))
				}
				out.RawByte(']')
			}
		}
	}
	{
		const prefix string = ",\"Logs\":"
		out.RawString(prefix)
		if in.Logs == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v18, v19 := range in.Logs {
				if v18 > 0 {
					out.RawByte(',')
				}
				out.String(string(v19))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"TokenAccountMap\":"
		out.RawString(prefix)
		if in.TokenAccountMap == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v20First := true
			for v20Name, v20Value := range in.TokenAccountMap {
				if v20First {
					v20First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v20Name))
				out.RawByte(':')
				(v20Value).MarshalEasyJSON(out)
			}
			out.RawByte('}')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ProcessInstructionData) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson89aae3efEncodeBlocsyInternalTypes6(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ProcessInstructionData) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson89aae3efEncodeBlocsyInternalTypes6(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ProcessInstructionData) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson89aae3efDecodeBlocsyInternalTypes6(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ProcessInstructionData) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson89aae3efDecodeBlocsyInternalTypes6(l, v)
}
func easyjson89aae3efDecodeBlocsyInternalTypes7(in *jlexer.Lexer, out *Pair) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "pair":
			out.Address = string(in.String())
		case "network":
			out.Network = string(in.String())
		case "exchange":
			out.Exchange = string(in.String())
		case "token":
			out.Token = string(in.String())
		case "quoteToken":
			(out.QuoteToken).UnmarshalEasyJSON(in)
		case "createdBlock":
			out.CreatedBlock = int64(in.Int64())
		case "createdTimestamp":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.CreatedTimestamp).UnmarshalJSON(data))
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson89aae3efEncodeBlocsyInternalTypes7(out *jwriter.Writer, in Pair) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"pair\":"
		out.RawString(prefix[1:])
		out.String(string(in.Address))
	}
	{
		const prefix string = ",\"network\":"
		out.RawString(prefix)
		out.String(string(in.Network))
	}
	{
		const prefix string = ",\"exchange\":"
		out.RawString(prefix)
		out.String(string(in.Exchange))
	}
	{
		const prefix string = ",\"token\":"
		out.RawString(prefix)
		out.String(string(in.Token))
	}
	{
		const prefix string = ",\"quoteToken\":"
		out.RawString(prefix)
		(in.QuoteToken).MarshalEasyJSON(out)
	}
	{
		const prefix string = ",\"createdBlock\":"
		out.RawString(prefix)
		out.Int64(int64(in.CreatedBlock))
	}
	{
		const prefix string = ",\"createdTimestamp\":"
		out.RawString(prefix)
		out.Raw((in.CreatedTimestamp).MarshalJSON())
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Pair) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson89aae3efEncodeBlocsyInternalTypes7(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Pair) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson89aae3efEncodeBlocsyInternalTypes7(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Pair) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson89aae3efDecodeBlocsyInternalTypes7(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Pair) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson89aae3efDecodeBlocsyInternalTypes7(l, v)
}
func easyjson89aae3efDecodeBlocsyInternalTypes8(in *jlexer.Lexer, out *BalanceSheet) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "Wallet":
			out.Wallet = string(in.String())
		case "Token":
			out.Token = string(in.String())
		case "Amount":
			out.Amount = float64(in.Float64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson89aae3efEncodeBlocsyInternalTypes8(out *jwriter.Writer, in BalanceSheet) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"Wallet\":"
		out.RawString(prefix[1:])
		out.String(string(in.Wallet))
	}
	{
		const prefix string = ",\"Token\":"
		out.RawString(prefix)
		out.String(string(in.Token))
	}
	{
		const prefix string = ",\"Amount\":"
		out.RawString(prefix)
		out.Float64(float64(in.Amount))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v BalanceSheet) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson89aae3efEncodeBlocsyInternalTypes8(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v BalanceSheet) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson89aae3efEncodeBlocsyInternalTypes8(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *BalanceSheet) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson89aae3efDecodeBlocsyInternalTypes8(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *BalanceSheet) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson89aae3efDecodeBlocsyInternalTypes8(l, v)
}
