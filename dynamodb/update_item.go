package dynamodb

import simplejson "github.com/bitly/go-simplejson"

/*
Construct an update item query.

The query can be composed via chaining and then executed via Execute()

Usage:
	update := table.UpdateItem(key)
			.ReturnValues(dynamodb.UPDATED_NEW)
			.UpdateExpression("SET Counter = Counter + :incr")
			.UpdateCondition("Counter < :checkVal")
			.ExpressionAttributes(NewNumberAttribute(":incr", "1"), NewNumberAttribute(":checkVal", 42))
	result, err := update.Execute()
	if err == nil {
		log.Printf("Counter is now %v", result.Attributes["Counter"].Value)
	}

*/
func (t *Table) UpdateItem(key *Key) *UpdateItem {
	q := NewQuery(t)
	q.AddKey(t, key)
	return &UpdateItem{table: t, query: q}
}

type UpdateItem struct {
	table               *Table
	query               *Query
	hasReturnValues     bool
	hasConsumedCapacity bool
}

func (u *UpdateItem) String() string {
	return u.query.String()
}

// Specify how return values are to be provided.
func (u *UpdateItem) ReturnValues(returnValues ReturnValues) *UpdateItem {
	u.hasReturnValues = (returnValues != NONE)
	u.query.AddReturnValues(returnValues)
	return u
}

// Specify whether consumed capacity is to be returned
func (u *UpdateItem) ReturnConsumedCapacity(shouldReturnConsumedCapacity bool) *UpdateItem {
	u.hasConsumedCapacity = shouldReturnConsumedCapacity
	u.query.ReturnConsumedCapacity(shouldReturnConsumedCapacity)
	return u
}

/*
Specify an update expression and optional attribute settings at the same time.

	update.UpdateExpression("SET Foo = Foo + :incr", dynamodb.NewNumberAttribute(":incr", "7"))

is equivalent to

	update.UpdateExpression("SET Foo = Foo + :incr")
	      .ExpressionAttributes(NewNumberAttribute(":incr", "7"))

*/
func (u *UpdateItem) UpdateExpression(expression string, attributes ...Attribute) *UpdateItem {
	u.query.AddUpdateExpression(expression)
	u.ExpressionAttributes(attributes...)
	return u
}

// Specify attribute value substitutions to be used in expressions.
func (u *UpdateItem) ExpressionAttributes(attributes ...Attribute) *UpdateItem {
	u.query.AddExpressionAttributes(attributes)
	return u
}

// Specify attribute name substitutions to be used in expressions.
func (u *UpdateItem) ExpressionAttributeNames(names map[string]string) *UpdateItem {
	u.query.AddExpressionAttributeNames(names)
	return u
}

// Specify a check condition for conditional updates.
func (u *UpdateItem) ConditionExpression(expression string) *UpdateItem {
	u.query.AddConditionExpression(expression)
	return u
}

// Execute this query and return complete json response
func (u *UpdateItem) Execute2() (*simplejson.Json, *UpdateResult, error) {
	jsonResponse, err := u.table.Server.queryServer(target("UpdateItem"), u.query)

	if err != nil {
		return nil, nil, err
	}

	var resp *simplejson.Json
	if u.hasConsumedCapacity || u.hasReturnValues {
		resp, err = simplejson.NewJson(jsonResponse)
		if err != nil {
			return resp, nil, err
		}
	}

	if resp != nil && u.hasReturnValues {
		attrib, err := resp.Get("Attributes").Map()
		if err != nil {
			return resp, nil, err
		}
		return resp, &UpdateResult{parseAttributes(attrib)}, nil
	}
	return resp, nil, nil
}

// Execute this query.
func (u *UpdateItem) Execute() (*UpdateResult, error) {
	_, result, err := u.Execute2()
	return result, err
}

type UpdateResult struct {
	Attributes map[string]*Attribute
}
