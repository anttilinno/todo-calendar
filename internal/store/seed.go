package store

// defaultTemplate holds name and content for a seed template.
type defaultTemplate struct {
	Name    string
	Content string
}

// defaultTemplates returns the pre-built templates seeded on first launch.
func defaultTemplates() []defaultTemplate {
	return []defaultTemplate{
		{Name: "Bug Report", Content: bugReportContent},
		{Name: "Checklist", Content: checklistContent},
		{Name: "Code Review", Content: codeReviewContent},
		{Name: "Daily Plan", Content: dailyPlanContent},
		{Name: "Feature Spec", Content: featureSpecContent},
		{Name: "Meeting Notes", Content: meetingNotesContent},
		{Name: "PR Checklist", Content: prChecklistContent},
	}
}

const meetingNotesContent = `## {{.Topic}}

**Date:** {{.Date}}
**Attendees:**

### Agenda
-

### Notes
-

### Action Items
- [ ] `

const checklistContent = `## {{.Title}}

- [ ]
- [ ]
- [ ]
- [ ]
- [ ] `

const dailyPlanContent = `## Daily Plan

### Top Priorities
1.
2.
3.

### Tasks
- [ ]
- [ ]
- [ ]

### Notes
`

const bugReportContent = `## Bug: {{.Summary}}

**Component:** {{.Component}}

### Steps to Reproduce
1.
2.
3.

### Expected Behavior


### Actual Behavior


### Environment
- OS:
- Version: `

const featureSpecContent = `## Feature: {{.Name}}

### Problem
What problem does this solve?

### Solution
How should it work?

### Acceptance Criteria
- [ ]
- [ ]
- [ ]

### Notes
`

const prChecklistContent = `## PR: {{.Title}}

### Changes
-

### Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes
- [ ] Self-reviewed code

### Testing
How was this tested?
`

const codeReviewContent = `## Review: {{.PR}}

**Author:** {{.Author}}

### Summary


### Feedback
-

### Approval
- [ ] Code quality
- [ ] Test coverage
- [ ] Documentation
`
