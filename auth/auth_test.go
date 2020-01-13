/*
Copyright © 2019, 2020 Red Hat, Inc.

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

package auth

import (
	"crypto/tls"
	"testing"
	"time"

	"gopkg.in/ldap.v3"
)

const testingDN = "uid=tester,ou=users,dc=test,dc=com"

type MockLdapClient struct {
	MemberOf *ldap.EntryAttribute
}

func (mlc *MockLdapClient) Start() {
	//noop
}

func (mlc *MockLdapClient) StartTLS(config *tls.Config) error {
	return nil
}

func (mlc *MockLdapClient) Close() {
	//noop
}

func (mlc *MockLdapClient) SetTimeout(t time.Duration) {
	//noop
}

func (mlc *MockLdapClient) Bind(username, password string) error {
	return nil
}

func (mlc *MockLdapClient) UnauthenticatedBind(username string) error {
	return nil
}

func (mlc *MockLdapClient) SimpleBind(r *ldap.SimpleBindRequest) (*ldap.SimpleBindResult, error) {
	return nil, nil
}

func (mlc *MockLdapClient) ExternalBind() error {
	return nil
}

func (mlc *MockLdapClient) Add(addReq *ldap.AddRequest) error {
	return nil
}

func (mlc *MockLdapClient) Del(delReq *ldap.DelRequest) error {
	return nil
}

func (mlc *MockLdapClient) Modify(modReq *ldap.ModifyRequest) error {
	return nil
}

func (mlc *MockLdapClient) ModifyDN(modReq *ldap.ModifyDNRequest) error {
	return nil
}

func (mlc *MockLdapClient) Compare(dn, attribute, value string) (bool, error) {
	return true, nil
}

func (mlc *MockLdapClient) PasswordModify(pmd *ldap.PasswordModifyRequest) (*ldap.PasswordModifyResult, error) {
	return nil, nil
}

func (mlc *MockLdapClient) Search(sr *ldap.SearchRequest) (*ldap.SearchResult, error) {
	entries := make([]*ldap.Entry, 1)
	attributes := make([]*ldap.EntryAttribute, 1)

	attributes[0] = mlc.MemberOf
	entries[0] = &ldap.Entry{
		DN:         testingDN,
		Attributes: attributes,
	}
	// mock SearchResult
	res := &ldap.SearchResult{
		Entries:   entries,
		Referrals: []string{},
		Controls:  []ldap.Control{},
	}
	return res, nil
}

func (mlc *MockLdapClient) SearchWithPaging(sr *ldap.SearchRequest, pagingSize uint32) (*ldap.SearchResult, error) {
	return nil, nil
}

func TestNoAccessLdapAuth(t *testing.T) {
	mockClient := &MockLdapClient{
		MemberOf: ldap.NewEntryAttribute("memberOf", []string{"dc=com", "dc=test", "ou=users"}),
	}
	_, err := ldapAuth("tester", "password", mockClient)
	if err.Error() != NoAccessMessage {
		t.Errorf("Expected error output: %s, but got %s", NoAccessMessage, err.Error())
	}
}

func TestSuccessfulLdapAuth(t *testing.T) {
	mockClient := &MockLdapClient{
		MemberOf: ldap.NewEntryAttribute("memberOf", []string{"dc=com", "dc=test", "ou=users", "cn=ccx-dev"}),
	}
	_, err := ldapAuth("tester", "password", mockClient)
	if err != nil {
		t.Error(err.Error())
	}
}
