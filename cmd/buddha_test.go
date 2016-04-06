package main

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/pusher/buddha"
)

// This module is mainly for testing the logic of how the results of Checks
// are handled

// TODO would also be nice to mock the actual command
// TODO would also be nice to mock the timeouts

const (
	DefaultFailures = 2 // Default number of times Checks should be retried
	DefaultDuration = buddha.Duration(10 ^ 9)
)

// MOCK CHECKS
// A mock check taht allows the results of the check are defined up front

type MockCheck struct {
	Str           string
	Results       []error
	TimesExecuted int
}

func (mockCheck *MockCheck) String() string {
	return mockCheck.Str
}

func (mockCheck *MockCheck) Validate() error {
	return errors.New("Did not expect Validate() to be called")
}

func (mockCheck *MockCheck) Execute(timeout time.Duration) error {
	mockCheck.TimesExecuted++
	if len(mockCheck.Results) < 1 {
		panic("MockCheck.Execute(): You must specifiy the number of return " +
			"results equal to the number of times you Execute to be called!")
	}
	result := mockCheck.Results[0]
	mockCheck.Results = mockCheck.Results[1:]
	return result
}

type MockChecks []MockCheck

// Need to explicity cast a slice of concrete MockChecks to the Check interface
func (mockChecks *MockChecks) toChecks() *buddha.Checks {
	checks := make(buddha.Checks, len(*mockChecks))
	for i := range *mockChecks {
		checks[i] = buddha.Check(&((*mockChecks)[i]))
	}
	return &checks
}

// JOB BUILDERS
// Functions for making the the constituation structs of a buddha.job that
// reference the mock Checks

func mkCheckReturning(results []error) MockCheck {
	return MockCheck{
		Str:           "mockCheck",
		Results:       results,
		TimesExecuted: 0,
	}
}

// Make checks by providing a list of result values, each time they are run, the next result will be returned
func mkChecksReturning(allResults ...[]error) MockChecks {
	checks := make([]MockCheck, len(allResults))
	for i, results := range allResults {
		checks[i] = mkCheckReturning(results)
	}
	return checks
}

// Specialisation of mkChecksReturning that makes checks that should only be run once
func mkChecksReturningOnce(allResults ...error) MockChecks {
	checks := make([]MockCheck, len(allResults))
	for i, result := range allResults {
		checks[i] = mkCheckReturning([]error{result})
	}
	return checks
}

func mkCommand(
	necessityChecks *buddha.Checks,
	beforeChecks *buddha.Checks,
	afterChecks *buddha.Checks,
	failures int,
	shouldSucceed bool,
) buddha.Command {
	name := "failing command"
	path := "false" // false unix command
	if shouldSucceed {
		name = "succeeding command"
		path = "true"
	}

	return buddha.Command{
		Name:      name,
		Path:      path,
		Args:      nil,
		Necessity: *necessityChecks,
		Before:    *beforeChecks,
		After:     *afterChecks,
		Grace:     DefaultDuration,
		Timeout:   DefaultDuration,
		Interval:  DefaultDuration,
		Failures:  failures,
		Stdout:    func(s string) { fmt.Print(s) },
	}
}

func mkJob(commands *[]buddha.Command) buddha.Job {
	return buddha.Job{
		Name:     "mock job",
		Root:     false, // Don't care
		Commands: *commands,
	}
}

// HELPERS FOR RUNNING JOBS
// Helpers for running the jobs and making assertions

// Return a check where the necessity checks return the values specified in the
// `returning` parameter, then check the necessity checks are run the expected
// number of times
func testNecessityCheck(t *testing.T, returning [][]error, timesExecuted []int, failures int, checkShouldContinue bool) {
	necessityMockChecks := mkChecksReturning(returning...)
	necessityChecks := necessityMockChecks.toChecks()
	beforeMockChecks := mkChecksReturningOnce(nil)
	beforeChecks := beforeMockChecks.toChecks()
	afterMockChecks := mkChecksReturningOnce(nil)
	afterChecks := afterMockChecks.toChecks()
	command := mkCommand(necessityChecks, beforeChecks, afterChecks, failures, true)
	commands := []buddha.Command{command}
	job := mkJob(&commands)

	runJob(&job)

	for i, results := range returning {
		if necessityMockChecks[i].TimesExecuted != len(results) {
			t.Fatalf("expected necessityMockChecks[%d] to be executed %d times", i, len(results))
		}
	}

	if checkShouldContinue {
		if beforeMockChecks[0].TimesExecuted != 1 {
			t.Fatal("expected beforeMockChecks[0] to be executed 1 time")
		}
		if afterMockChecks[0].TimesExecuted != 1 {
			t.Fatal("expected afterMockChecks[0] to be executed 1 time")
		}
	} else {
		if beforeMockChecks[0].TimesExecuted != 0 {
			t.Fatal("expected beforeMockChecks[0] to be executed 0 times")
		}
		if afterMockChecks[0].TimesExecuted != 0 {
			t.Fatal("expected afterMockChecks[0] to be executed 0 times")
		}
	}
}

func testBeforeCheck(t *testing.T, returning [][]error, timesExecuted []int, failures int, checkShouldContinue bool) {
	necessityMockChecks := mkChecksReturningOnce(nil)
	necessityChecks := necessityMockChecks.toChecks()
	beforeMockChecks := mkChecksReturning(returning...)
	beforeChecks := beforeMockChecks.toChecks()
	afterMockChecks := mkChecksReturningOnce(nil)
	afterChecks := afterMockChecks.toChecks()
	command := mkCommand(necessityChecks, beforeChecks, afterChecks, failures, true)
	commands := []buddha.Command{command}
	job := mkJob(&commands)

	runJob(&job)

	if necessityMockChecks[0].TimesExecuted != 1 {
		t.Fatal("expected necessityMockChecks[0] to be executed 1 time")
	}

	for i, results := range returning {
		if beforeMockChecks[i].TimesExecuted != len(results) {
			t.Fatalf("expected beforeMockChecks[%d] to be executed %d times", i, len(results))
		}
	}

	if checkShouldContinue {
		if afterMockChecks[0].TimesExecuted != 1 {
			t.Fatal("expected afterMockChecks[0] to be executed 1 time")
		}
	} else {
		if afterMockChecks[0].TimesExecuted != 0 {
			t.Fatal("expected afterMockChecks[0] to be executed 0 times")
		}
	}
}

func testAfterCheck(t *testing.T, returning [][]error, timesExecuted []int, failures int) {
	necessityMockChecks := mkChecksReturningOnce(nil)
	necessityChecks := necessityMockChecks.toChecks()
	beforeMockChecks := mkChecksReturningOnce(nil)
	beforeChecks := beforeMockChecks.toChecks()
	afterMockChecks := mkChecksReturning(returning...)
	afterChecks := afterMockChecks.toChecks()
	command := mkCommand(necessityChecks, beforeChecks, afterChecks, failures, true)
	commands := []buddha.Command{command}
	job := mkJob(&commands)

	runJob(&job)

	if necessityMockChecks[0].TimesExecuted != 1 {
		t.Fatal("expected necessityMockChecks[0] to be executed 1 time")
	}
	if beforeMockChecks[0].TimesExecuted != 1 {
		t.Fatal("expected beforeMockChecks[0] to be executed 1 time")
	}

	for i, results := range returning {
		if afterMockChecks[i].TimesExecuted != len(results) {
			t.Fatalf("expected afterMockChecks[%d] to be executed %d times", i, len(results))
		}
	}
}

// THE TESTS

// NECESSITY CHECKS

func TestRunJobNecessary(t *testing.T) {
	testNecessityCheck(
		t,
		[][]error{{nil}},
		[]int{1},
		DefaultFailures,
		true,
	)
}

func TestRunJobUnecessary(t *testing.T) {
	testNecessityCheck(
		t,
		[][]error{{buddha.CheckFalse("dummy false")}},
		[]int{1},
		DefaultFailures,
		false,
	)
}

func TestRunJobSomeUnecessary(t *testing.T) {
	testNecessityCheck(
		t,
		[][]error{{nil}, {buddha.CheckFalse("dummy false")}},
		[]int{1, 1},
		DefaultFailures,
		true,
	)
}

func TestRunJobNecessityError(t *testing.T) {
	testNecessityCheck(
		t,
		[][]error{{errors.New("error"), errors.New("error")}},
		[]int{DefaultFailures},
		DefaultFailures,
		false,
	)
}

func TestRunJobNecessitySomeError(t *testing.T) {
	testNecessityCheck(
		t,
		[][]error{{nil}, {errors.New("error"), errors.New("error")}},
		[]int{1, DefaultFailures},
		DefaultFailures,
		false,
	)
}

func TestRunJobNecessityErrorThenSucceed(t *testing.T) {
	testNecessityCheck(
		t,
		[][]error{{errors.New("error"), nil}},
		[]int{DefaultFailures},
		DefaultFailures,
		true,
	)
}

// BEFORE CHECKS

func TestRunJobBeforeCheckTrue(t *testing.T) {
	testBeforeCheck(
		t,
		[][]error{{nil}},
		[]int{1},
		DefaultFailures,
		true,
	)
}

func TestRunJobBeforeCheckFalse(t *testing.T) {
	testBeforeCheck(
		t,
		[][]error{{buddha.CheckFalse("dummy false"), buddha.CheckFalse("dummy false")}},
		[]int{DefaultFailures},
		DefaultFailures,
		false,
	)
}

func TestRunJobBeforeCheckSomeFalse(t *testing.T) {
	testBeforeCheck(
		t,
		[][]error{{nil}, {buddha.CheckFalse("dummy false"), buddha.CheckFalse("dummy false")}},
		[]int{1, DefaultFailures},
		DefaultFailures,
		false,
	)
}

func TestRunJobBeforeCheckFalseThenTrue(t *testing.T) {
	testBeforeCheck(
		t,
		[][]error{{buddha.CheckFalse("dummy false"), nil}},
		[]int{DefaultFailures},
		DefaultFailures,
		true,
	)
}

func TestRunJobBeforeCheckError(t *testing.T) {
	testBeforeCheck(
		t,
		[][]error{{errors.New("error"), errors.New("error")}},
		[]int{DefaultFailures},
		DefaultFailures,
		false,
	)
}

func TestRunJobBeforeCheckSomeError(t *testing.T) {
	testBeforeCheck(
		t,
		[][]error{{nil}, {errors.New("error"), errors.New("error")}},
		[]int{1, DefaultFailures},
		DefaultFailures,
		false,
	)
}

func TestRunJobBeforeCheckErrorThenSucceed(t *testing.T) {
	testBeforeCheck(
		t,
		[][]error{{errors.New("error"), nil}},
		[]int{DefaultFailures},
		DefaultFailures,
		true,
	)
}

// AFTER CHECKS

func TestRunJobAfterCheckTrue(t *testing.T) {
	testAfterCheck(
		t,
		[][]error{{nil}},
		[]int{1},
		DefaultFailures,
	)
}

func TestRunJobAfterCheckFalse(t *testing.T) {
	testAfterCheck(
		t,
		[][]error{{buddha.CheckFalse("dummy false"), buddha.CheckFalse("dummy false")}},
		[]int{DefaultFailures},
		DefaultFailures,
	)
}

func TestRunJobAfterCheckSomeFalse(t *testing.T) {
	testAfterCheck(
		t,
		[][]error{{nil}, {buddha.CheckFalse("dummy false"), buddha.CheckFalse("dummy false")}},
		[]int{1, DefaultFailures},
		DefaultFailures,
	)
}

func TestRunJobAfterCheckFalseThenTrue(t *testing.T) {
	testAfterCheck(
		t,
		[][]error{{buddha.CheckFalse("dummy false"), nil}},
		[]int{DefaultFailures},
		DefaultFailures,
	)
}

func TestRunJobAfterCheckError(t *testing.T) {
	testAfterCheck(
		t,
		[][]error{{errors.New("error"), errors.New("error")}},
		[]int{DefaultFailures},
		DefaultFailures,
	)
}

func TestRunJobAfterCheckSomeError(t *testing.T) {
	testAfterCheck(
		t,
		[][]error{{nil}, {errors.New("error"), errors.New("error")}},
		[]int{1, DefaultFailures},
		DefaultFailures,
	)
}

func TestRunJobAfterCheckErrorThenSucceed(t *testing.T) {
	testAfterCheck(
		t,
		[][]error{{errors.New("error"), nil}},
		[]int{DefaultFailures},
		DefaultFailures,
	)
}

// FAILING COMMAND

func TestRunJobAfterChecksNotRunIfCommandFails(t *testing.T) {
	necessityMockChecks := mkChecksReturningOnce(nil)
	necessityChecks := necessityMockChecks.toChecks()
	beforeMockChecks := mkChecksReturningOnce(nil)
	beforeChecks := beforeMockChecks.toChecks()
	afterMockChecks := mkChecksReturningOnce(nil)
	afterChecks := afterMockChecks.toChecks()
	// Make a command that fails
	command := mkCommand(necessityChecks, beforeChecks, afterChecks, DefaultFailures, false)
	commands := []buddha.Command{command}
	job := mkJob(&commands)

	runJob(&job)

	if necessityMockChecks[0].TimesExecuted != 1 {
		t.Fatal("expected necessityMockChecks[0] to be executed 1 time")
	}
	if beforeMockChecks[0].TimesExecuted != 1 {
		t.Fatal("expected beforeMockChecks[0] to be executed 1 time")
	}
	if afterMockChecks[0].TimesExecuted != 0 {
		t.Fatal("expected afterMockChecks[0] to be executed 0 times")
	}
}

// MULTIPLE COMMANDS

func TestRunJobMultipleCommands(t *testing.T) {
	necessityMockChecks1 := mkChecksReturningOnce(nil)
	necessityChecks1 := necessityMockChecks1.toChecks()
	beforeMockChecks1 := mkChecksReturningOnce(nil)
	beforeChecks1 := beforeMockChecks1.toChecks()
	afterMockChecks1 := mkChecksReturningOnce(nil)
	afterChecks1 := afterMockChecks1.toChecks()
	command1 := mkCommand(necessityChecks1, beforeChecks1, afterChecks1, DefaultFailures, true)

	necessityMockChecks2 := mkChecksReturningOnce(nil)
	necessityChecks2 := necessityMockChecks2.toChecks()
	beforeMockChecks2 := mkChecksReturningOnce(nil)
	beforeChecks2 := beforeMockChecks2.toChecks()
	afterMockChecks2 := mkChecksReturningOnce(nil)
	afterChecks2 := afterMockChecks2.toChecks()
	command2 := mkCommand(necessityChecks2, beforeChecks2, afterChecks2, DefaultFailures, true)

	commands := []buddha.Command{command1, command2}
	job := mkJob(&commands)

	runJob(&job)

	if necessityMockChecks1[0].TimesExecuted != 1 {
		t.Fatal("expected necessityMockChecks1[0] to be executed 1 time")
	}
	if beforeMockChecks1[0].TimesExecuted != 1 {
		t.Fatal("expected beforeMockChecks1[0] to be executed 1 time")
	}
	if afterMockChecks1[0].TimesExecuted != 1 {
		t.Fatal("expected afterMockChecks1[0] to be executed 1 times")
	}

	if necessityMockChecks2[0].TimesExecuted != 1 {
		t.Fatal("expected necessityMockChecks2[0] to be executed 1 time")
	}
	if beforeMockChecks2[0].TimesExecuted != 1 {
		t.Fatal("expected beforeMockChecks2[0] to be executed 1 time")
	}
	if afterMockChecks2[0].TimesExecuted != 1 {
		t.Fatal("expected afterMockChecks2[0] to be executed 1 times")
	}
}

func TestRunJobMultipleCommandsFirstErrors(t *testing.T) {
	necessityMockChecks1 := mkChecksReturningOnce(nil)
	necessityChecks1 := necessityMockChecks1.toChecks()
	beforeMockChecks1 := mkChecksReturningOnce(nil)
	beforeChecks1 := beforeMockChecks1.toChecks()
	// Error in the after check
	afterMockChecks1 := mkChecksReturning([]error{errors.New("error"), errors.New("error")})
	afterChecks1 := afterMockChecks1.toChecks()
	command1 := mkCommand(necessityChecks1, beforeChecks1, afterChecks1, DefaultFailures, true)

	// Expect this to not be run
	necessityMockChecks2 := mkChecksReturningOnce(nil)
	necessityChecks2 := necessityMockChecks2.toChecks()
	beforeMockChecks2 := mkChecksReturningOnce(nil)
	beforeChecks2 := beforeMockChecks2.toChecks()
	afterMockChecks2 := mkChecksReturningOnce(nil)
	afterChecks2 := afterMockChecks2.toChecks()
	command2 := mkCommand(necessityChecks2, beforeChecks2, afterChecks2, DefaultFailures, true)

	commands := []buddha.Command{command1, command2}
	job := mkJob(&commands)

	runJob(&job)

	if necessityMockChecks1[0].TimesExecuted != 1 {
		t.Fatal("expected necessityMockChecks1[0] to be executed 1 time")
	}
	if beforeMockChecks1[0].TimesExecuted != 1 {
		t.Fatal("expected beforeMockChecks1[0] to be executed 1 time")
	}
	if afterMockChecks1[0].TimesExecuted != DefaultFailures {
		t.Fatalf("expected afterMockChecks1[0] to be executed %d times", DefaultFailures)
	}

	if necessityMockChecks2[0].TimesExecuted != 0 {
		t.Fatal("expected necessityMockChecks2[0] to be executed 0 time")
	}
	if beforeMockChecks2[0].TimesExecuted != 0 {
		t.Fatal("expected beforeMockChecks2[0] to be executed 0 time")
	}
	if afterMockChecks2[0].TimesExecuted != 0 {
		t.Fatal("expected afterMockChecks2[0] to be executed 0 times")
	}
}
