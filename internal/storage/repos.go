package storage

// Repos aggregates all repository types for dependency injection.
type Repos struct {
	Projects  *ProjectRepo
	Documents *DocumentRepo
	Outputs   *OutputRepo
	Sessions  *SessionRepo
	Context   *ContextRepo
}

// NewRepos creates all repositories from an open DB.
func NewRepos(db *DB) *Repos {
	projects := newProjectRepo(db.db)
	outputs := newOutputRepo(db.db)
	sessions := newSessionRepo(db.db)
	return &Repos{
		Projects:  projects,
		Documents: newDocumentRepo(db.db),
		Outputs:   outputs,
		Sessions:  sessions,
		Context:   newContextRepo(db.db, projects, outputs, sessions),
	}
}
