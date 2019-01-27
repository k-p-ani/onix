Feature: Create a new Link Rule
  As an API user
  I want to create a new Link Rule
  So that I can restrict the types of items that can be linked

  Scenario: Create a new Link Rule
    Given the link rule does not exist in the database
    Given the link rule URL of the service with key is known
    Given the link rule natural key is known
    Given a json payload with new link rule information exists
    When a link rule PUT HTTP request with a JSON payload is done
    Then the response code is 200
    Then the response has body