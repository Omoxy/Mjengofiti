package inspections_test

import (
	"context"
	"testing"

	"github.com/mjengofiti/api/internal/inspections"
)

func TestInspectionStateMachineUsesAssignmentOperation(t *testing.T) {
	repository := inspections.NewRepository(nil)
	ctx := context.Background()
	caseRecord := &inspections.Case{
		BuildingID:     "building-test-001",
		Title:          "Test case",
		Priority:       inspections.PriorityHigh,
		ReferralReason: "Test fixture",
		CreatedBy:      "test-supervisor",
	}

	if err := repository.Create(ctx, caseRecord); err != nil {
		t.Fatalf("create case: %v", err)
	}
	if err := repository.UpdateStatus(ctx, caseRecord.ID, inspections.StatusInProgress, "test-supervisor"); err == nil {
		t.Fatal("expected open case to reject direct transition to in progress")
	}
	if err := repository.UpdateStatus(ctx, caseRecord.ID, inspections.StatusTriaged, "test-supervisor"); err != nil {
		t.Fatalf("triage case: %v", err)
	}
	if err := repository.UpdateStatus(ctx, caseRecord.ID, inspections.StatusAssigned, "test-supervisor"); err == nil {
		t.Fatal("expected generic status update to reject assignment without an assignee")
	}
	if err := repository.Assign(ctx, caseRecord.ID, "test-inspector", "test-supervisor"); err != nil {
		t.Fatalf("assign case: %v", err)
	}
	if caseRecord.AssignedTo != nil {
		t.Fatal("repository assignment must not mutate the caller-owned case value")
	}
	assigned, err := repository.FindByID(ctx, caseRecord.ID)
	if err != nil {
		t.Fatalf("read assigned case: %v", err)
	}
	if assigned.Status != inspections.StatusAssigned || assigned.AssignedTo == nil || *assigned.AssignedTo != "test-inspector" {
		t.Fatalf("assignment was not persisted: %#v", assigned)
	}
	if err := repository.UpdateStatus(ctx, caseRecord.ID, inspections.StatusInProgress, "test-inspector"); err != nil {
		t.Fatalf("start assigned case: %v", err)
	}

	history, err := repository.GetHistory(ctx, caseRecord.ID)
	if err != nil {
		t.Fatalf("read case history: %v", err)
	}
	if len(history) != 4 {
		t.Fatalf("expected create, triage, assignment, and start history entries, got %d", len(history))
	}
}

func TestListAssignedRestrictsInspectorCases(t *testing.T) {
	repository := inspections.NewRepository(nil)
	ctx := context.Background()

	for _, assignee := range []string{"inspector-a", "inspector-b"} {
		caseRecord := &inspections.Case{BuildingID: "building-test", Title: "Test case", CreatedBy: "test-supervisor"}
		if err := repository.Create(ctx, caseRecord); err != nil {
			t.Fatalf("create case: %v", err)
		}
		if err := repository.UpdateStatus(ctx, caseRecord.ID, inspections.StatusTriaged, "test-supervisor"); err != nil {
			t.Fatalf("triage case: %v", err)
		}
		if err := repository.Assign(ctx, caseRecord.ID, assignee, "test-supervisor"); err != nil {
			t.Fatalf("assign case: %v", err)
		}
	}

	cases, total, err := repository.ListAssigned(ctx, nil, "inspector-a", 50, 0)
	if err != nil {
		t.Fatalf("list assigned cases: %v", err)
	}
	if total != 1 || len(cases) != 1 || cases[0].AssignedTo == nil || *cases[0].AssignedTo != "inspector-a" {
		t.Fatalf("inspector received cases assigned to another user: total=%d cases=%#v", total, cases)
	}
}

func TestIsValidTransition(t *testing.T) {
	for _, testCase := range []struct {
		from     string
		to       string
		expected bool
	}{
		{inspections.StatusOpen, inspections.StatusTriaged, true},
		{inspections.StatusOpen, inspections.StatusClosed, true},
		{inspections.StatusOpen, inspections.StatusInProgress, false},
		{inspections.StatusTriaged, inspections.StatusAssigned, true},
		{inspections.StatusAssigned, inspections.StatusInProgress, true},
		{inspections.StatusInProgress, inspections.StatusPendingReview, true},
		{inspections.StatusPendingReview, inspections.StatusClosed, true},
		{inspections.StatusPendingReview, inspections.StatusReopened, true},
		{inspections.StatusClosed, inspections.StatusReopened, true},
		{inspections.StatusClosed, inspections.StatusInProgress, false},
	} {
		if actual := inspections.IsValidTransition(testCase.from, testCase.to); actual != testCase.expected {
			t.Errorf("transition %s -> %s: expected %v, got %v", testCase.from, testCase.to, testCase.expected, actual)
		}
	}
}
