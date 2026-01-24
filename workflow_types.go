// Package main - workflow type aliases
// This file re-exports types from pkg/types to maintain a single source of truth
// while preserving backward compatibility with existing code in main package.
package main

import (
	"Gaze/pkg/types"
)

// ============== Type Aliases ==============
// Re-export all workflow types from pkg/types

type StepConnections = types.StepConnections
type ElementSelector = types.ElementSelector
type StepCommon = types.StepCommon
type HandleInfo = types.HandleInfo
type StepLayout = types.StepLayout
type TapParams = types.TapParams
type SwipeParams = types.SwipeParams
type ElementParams = types.ElementParams
type AppParams = types.AppParams
type BranchParams = types.BranchParams
type WaitParams = types.WaitParams
type ScriptParams = types.ScriptParams
type VariableParams = types.VariableParams
type ADBParams = types.ADBParams
type SubWorkflowParams = types.SubWorkflowParams
type ReadToVariableParams = types.ReadToVariableParams
type SessionParams = types.SessionParams
type WorkflowStep = types.WorkflowStep
type Workflow = types.Workflow
type ValidationError = types.ValidationError
type WorkflowExecutionResult = types.WorkflowExecutionResult

// ============== Function Aliases ==============

// SerializeWorkflow serializes a workflow to JSON
var SerializeWorkflow = types.SerializeWorkflow

// ParseWorkflow parses workflow JSON
var ParseWorkflow = types.ParseWorkflow
