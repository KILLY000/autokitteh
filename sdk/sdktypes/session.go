package sdktypes

import (
	"errors"
	"time"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type Session struct {
	object[*SessionPB, SessionTraits]
}

func init() { registerObject[Session]() }

var InvalidSession Session

type SessionPB = sessionv1.Session

type SessionTraits struct{}

func (SessionTraits) Validate(m *SessionPB) error {
	return errors.Join(
		enumField[SessionStateType]("state", m.State),
		idField[BuildID]("build_id", m.BuildId),
		idField[ProjectID]("project_id", m.ProjectId),
		idField[DeploymentID]("deployment_id", m.DeploymentId),
		idField[EventID]("event_id", m.EventId),
		idField[SessionID]("parent_session_id", m.ParentSessionId),
		idField[SessionID]("session_id", m.SessionId),
		objectField[CodeLocation]("entrypoint", m.Entrypoint),
		valuesMapField("inputs", m.Inputs),
	)
}

func (SessionTraits) StrictValidate(m *SessionPB) error {
	return errors.Join(
		mandatory("entrypoint", m.Entrypoint),
		mandatory("build_id", m.BuildId),
		mandatory("project_id", m.ProjectId),
	)
}

func (SessionTraits) Mutables() []string { return []string{"state"} }

func SessionFromProto(m *SessionPB) (Session, error) { return FromProto[Session](m) }
func StrictSessionFromProto(m *SessionPB) (Session, error) {
	return Strict(SessionFromProto(m))
}

func (p Session) WithNewID() Session {
	return Session{p.forceUpdate(func(pb *SessionPB) { pb.SessionId = NewSessionID().String() })}
}

func (p Session) WithNoID() Session {
	return Session{p.forceUpdate(func(pb *SessionPB) { pb.SessionId = "" })}
}

func (p Session) ID() SessionID { return kittehs.Must1(ParseSessionID(p.read().SessionId)) }

func (p Session) DeploymentID() DeploymentID {
	return kittehs.Must1(ParseDeploymentID(p.read().DeploymentId))
}
func (p Session) EventID() EventID         { return kittehs.Must1(ParseEventID(p.read().EventId)) }
func (p Session) BuildID() BuildID         { return kittehs.Must1(ParseBuildID(p.read().BuildId)) }
func (p Session) ProjectID() ProjectID     { return kittehs.Must1(ParseProjectID(p.read().ProjectId)) }
func (p Session) EntryPoint() CodeLocation { return forceFromProto[CodeLocation](p.read().Entrypoint) }
func (p Session) Memo() map[string]string  { return p.read().Memo }
func (p Session) Inputs() map[string]Value {
	return kittehs.TransformMapValues(p.read().Inputs, forceFromProto[Value])
}
func (p Session) CreatedAt() time.Time { return p.read().CreatedAt.AsTime() }
func (p Session) ParentSessionID() SessionID {
	return kittehs.Must1(ParseSessionID(p.read().ParentSessionId))
}

func (p Session) State() SessionStateType {
	return forceEnumFromProto[SessionStateType](p.read().State)
}

func (p Session) WithInputs(inputs map[string]Value) Session {
	return Session{p.forceUpdate(func(pb *SessionPB) { pb.Inputs = kittehs.TransformMapValues(inputs, ToProto) })}
}

func NewSession(buildID BuildID, ep CodeLocation, inputs map[string]Value, memo map[string]string) Session {
	return kittehs.Must1(SessionFromProto(
		&SessionPB{
			BuildId:    buildID.String(),
			Entrypoint: ToProto(ep),
			Inputs:     kittehs.TransformMapValues(inputs, ToProto),
			Memo:       memo,
		},
	))
}

func (s Session) WithParentSessionID(id SessionID) Session {
	return Session{s.forceUpdate(func(pb *SessionPB) { pb.ParentSessionId = id.String() })}
}

func (s Session) WithDeploymentID(id DeploymentID) Session {
	return Session{s.forceUpdate(func(pb *SessionPB) { pb.DeploymentId = id.String() })}
}

func (s Session) WithProjectID(id ProjectID) Session {
	return Session{s.forceUpdate(func(pb *SessionPB) { pb.ProjectId = id.String() })}
}

func (s Session) WithEventID(id EventID) Session {
	return Session{s.forceUpdate(func(pb *SessionPB) { pb.EventId = id.String() })}
}

func (s Session) WithBuildID(id BuildID) Session {
	return Session{s.forceUpdate(func(pb *SessionPB) { pb.BuildId = id.String() })}
}

func (s Session) WithEndpoint(ep CodeLocation) Session {
	return Session{s.forceUpdate(func(pb *SessionPB) { pb.Entrypoint = ToProto(ep) })}
}

func (s Session) WithID(id SessionID) Session {
	return Session{s.forceUpdate(func(pb *SessionPB) { pb.SessionId = id.String() })}
}

func (s Session) WithState(state SessionStateType) Session {
	return Session{s.forceUpdate(func(pb *SessionPB) { pb.State = state.ToProto() })}
}
