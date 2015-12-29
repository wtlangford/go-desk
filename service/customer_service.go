package service

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	. "github.com/wtlangford/go-desk/resource"
)

type CustomerService struct {
	client *Client
}

// Get retrieves a customer.
// See Desk API: http://dev.desk.com/API/customers/#show
func (c *CustomerService) Get(id string) (*Customer, *http.Response, error) {
	restful := Restful{}
	customer := NewCustomer()
	path := NewIdentityResourcePath(id, customer)
	resp, err := restful.
		Get(path.Path()).
		Json(customer).
		Client(c.client).
		Do()
	return customer, resp, err
}

// List customers with filtering and pagination.
// See Desk API: http://dev.desk.com/API/customers/#list
func (c *CustomerService) List(params *url.Values) (*Page, *http.Response, error) {
	restful := Restful{}
	page := new(Page)
	path := NewResourcePath(NewCustomer())
	resp, err := restful.
		Get(path.Path()).
		Json(page).
		Params(params).
		Client(c.client).
		Do()
	if err != nil {
		return nil, resp, err
	}
	err = c.unravelPage(page)
	if err != nil {
		return nil, nil, err
	}
	return page, resp, err
}

// Search customers with filtering and pagination.
// See Desk API: http://dev.desk.com/API/customers/#search
func (c *CustomerService) Search(params *url.Values, q *string) (*Page, *http.Response, error) {
	restful := Restful{}
	page := new(Page)
	path := NewResourcePath(NewCustomer()).SetAction("search")
	resp, err := restful.
		Get(path.Path()).
		Json(page).
		Params(params).
		Client(c.client).
		Do()
	if err != nil {
		return nil, resp, err
	}
	err = c.unravelPage(page)
	if err != nil {
		return nil, nil, err
	}
	return page, resp, err
}

// Create a customer.
// See Desk API: http://dev.desk.com/API/customers/#create
func (c *CustomerService) Create(customer *Customer) (*Customer, *http.Response, error) {
	restful := Restful{}
	createdCustomer := new(Customer)
	path := NewResourcePath(NewCustomer())
	resp, err := restful.
		Post(path.Path()).
		Body(customer).
		Json(createdCustomer).
		Client(c.client).
		Do()
	return createdCustomer, resp, err
}

// Update a customer.
// See Desk API: http://dev.desk.com/API/customers/#update
func (c *CustomerService) Update(customer *Customer) (*Customer, *http.Response, error) {
	restful := Restful{}
	updatedCustomer := new(Customer)
	path := NewIdentityResourcePath(customer.GetResourceId(), customer)
	resp, err := restful.
		Patch(path.Path()).
		Body(customer).
		Json(updatedCustomer).
		Client(c.client).
		Do()
	return updatedCustomer, resp, err
}

type MergeOverrides struct {
	FirstName    *string                `json:"first_name,omitempty"`
	LastName     *string                `json:"last_name,omitempty"`
	Company      *string                `json:"company,omitempty"`
	Title        *string                `json:"title,omitempty"`
	Description  *string                `json:"description,omitempty"`
	ExternalID   *string                `json:"external_id,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	Hal
}

// Merge customers.
// See Desk API: http://dev.desk.com/API/customers/#merge
func (c *CustomerService) Merge(customer *Customer, overrides *MergeOverrides, customers ...*Customer) (string, *http.Response, error) {
	for _, c := range customers {
		overrides.AddHrefLink("customers", c.GetResourcePath(c).String())
	}
	restful := Restful{}
	hal := new(Hal)
	path := NewIdentityResourcePath(customer.GetResourceId(), NewCustomer()).SetAction("merge")
	resp, err := restful.
		Post(path.Path()).
		Body(overrides).
		Json(hal).
		Client(c.client).
		Do()

	var jobid string
	if err == nil {
		bits := strings.Split(hal.GetHrefLink("job"), "/")
		if len(bits) > 0 {
			jobid = bits[len(bits)-1]
		}
	}

	return jobid, resp, err
}

// Cases provides a list of cases associated with a customer.
// See Desk API: http://dev.desk.com/API/customers/#list-cases
func (c *CustomerService) Cases(id string, params *url.Values) (*Page, *http.Response, error) {
	restful := Restful{}
	page := new(Page)
	path := NewIdentityResourcePath(id, NewCustomer()).SetNested(NewCase())
	resp, err := restful.
		Get(path.Path()).
		Json(page).
		Params(params).
		Client(c.client).
		Do()
	if err != nil {
		return nil, resp, err
	}
	err = c.unravelPage(page)
	if err != nil {
		return nil, nil, err
	}
	return page, resp, err
}

func (c *CustomerService) unravelPage(page *Page) error {
	customers := new([]Customer)
	err := json.Unmarshal(*page.Embedded.RawEntries, &customers)
	if err != nil {
		return err
	}
	page.Embedded.Entries = make([]interface{}, len(*customers))
	for i, v := range *customers {
		v.InitializeResource(v)
		page.Embedded.Entries[i] = interface{}(v)
	}
	page.Embedded.RawEntries = nil
	return err
}
