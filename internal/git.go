package internal

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type DeveloperAliases map[string]string

type RepoCommitActivity struct {
	RepoName string
	Activity *CommitActivity
}

type CombinedCommitActivity struct {
	Repos []RepoCommitActivity
}

func (cca *CombinedCommitActivity) Add(repoName string, activity *CommitActivity) {
	cca.Repos = append(cca.Repos, RepoCommitActivity{RepoName: repoName, Activity: activity})
}

type CommitActivity struct {
	Weekdays map[string][]int // Developer -> Weekday activity
	Hours    map[string][]int // Developer -> Hour activity
	Months   map[string][]int // Developer -> Month activity
	Weeks    map[string][]int // Developer -> Week activity
}

func NewCommitActivity() *CommitActivity {
	return &CommitActivity{
		Weekdays: make(map[string][]int),
		Hours:    make(map[string][]int),
		Months:   make(map[string][]int),
		Weeks:    make(map[string][]int),
	}
}

func (ca *CommitActivity) AddActivity(developer string, weekday, hour, month, week, value int) {
	if _, exists := ca.Weekdays[developer]; !exists {
		ca.Weekdays[developer] = make([]int, 7)
	}
	ca.Weekdays[developer][weekday] += value

	if _, exists := ca.Hours[developer]; !exists {
		ca.Hours[developer] = make([]int, 24)
	}
	ca.Hours[developer][hour] += value

	if _, exists := ca.Months[developer]; !exists {
		ca.Months[developer] = make([]int, 12)
	}
	ca.Months[developer][month] += value

	if _, exists := ca.Weeks[developer]; !exists {
		ca.Weeks[developer] = make([]int, 53)
	}
	ca.Weeks[developer][week] += value
}

func (ca *CommitActivity) Combine(other *CommitActivity) {
	// Combine Weekdays
	for developer, data := range other.Weekdays {
		if _, exists := ca.Weekdays[developer]; !exists {
			ca.Weekdays[developer] = make([]int, len(data))
		}
		for i, value := range data {
			ca.Weekdays[developer][i] += value
		}
	}

	// Combine Hours
	for developer, data := range other.Hours {
		if _, exists := ca.Hours[developer]; !exists {
			ca.Hours[developer] = make([]int, len(data))
		}
		for i, value := range data {
			ca.Hours[developer][i] += value
		}
	}

	// Combine Months
	for developer, data := range other.Months {
		if _, exists := ca.Months[developer]; !exists {
			ca.Months[developer] = make([]int, len(data))
		}
		for i, value := range data {
			ca.Months[developer][i] += value
		}
	}

	// Combine Weeks
	for developer, data := range other.Weeks {
		if _, exists := ca.Weeks[developer]; !exists {
			ca.Weeks[developer] = make([]int, len(data))
		}
		for i, value := range data {
			ca.Weeks[developer][i] += value
		}
	}
}

func AnalyzeCommits(repoPath string, aliases DeveloperAliases, mode string) (*CommitActivity, error) {
	activity := NewCommitActivity()

	// Open the Git repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("could not open repository: %w", err)
	}

	// Get the commit history
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("could not get repository head: %w", err)
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve commits: %w", err)
	}

	// Iterate through the commits
	err = commitIter.ForEach(func(c *object.Commit) error {
		commitTime := c.Author.When
		authorEmail := strings.ToLower(c.Author.Email)

		// Map alias to developer name
		developer, exists := aliases[authorEmail]
		if !exists {
			developer = "Unknown"
		}

		// Weekday, hour, month, and week
		weekday := int(commitTime.Weekday())
		hour := commitTime.Hour()
		month := int(commitTime.Month()) - 1 // `time.Month` is 1-based
		_, week := commitTime.ISOWeek()

		if mode == "commits" {
			// Increment commit count for the developer
			activity.AddActivity(developer, weekday, hour, month, week, 1)
		} else if mode == "lines" {
			// Aggregate changed lines
			stats, err := c.Stats()
			if err != nil {
				return fmt.Errorf("could not get diff stats: %w", err)
			}

			// Sum added and deleted lines
			lineChanges := 0
			for _, stat := range stats {
				lineChanges += stat.Addition + stat.Deletion
			}

			// Increment line change count for the developer
			activity.AddActivity(developer, weekday, hour, month, week, lineChanges)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not iterate through commits: %w", err)
	}

	return activity, nil
}

func AnalyzeCommitsInRange(repoPath string, start, end time.Time, aliases DeveloperAliases) (*CommitActivity, error) {
	activity := NewCommitActivity()

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("could not open repository: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("could not get repository head: %w", err)
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve commits: %w", err)
	}

	err = commitIter.ForEach(func(c *object.Commit) error {
		commitTime := c.Author.When
		authorEmail := strings.ToLower(c.Author.Email)

		// Map alias to developer name
		developer, exists := aliases[authorEmail]
		if !exists {
			developer = "Unknown"
		}

		// Filter commits based on date range
		if (!start.IsZero() && commitTime.Before(start)) || (!end.IsZero() && commitTime.After(end)) {
			return nil
		}

		// Weekday and hour
		weekday := commitTime.Weekday()
		hour := commitTime.Hour()

		// Month (0 = January, 11 = December)
		month := commitTime.Month() - 1

		// Week number (0 = first week of year)
		_, week := commitTime.ISOWeek()

		activity.AddActivity(developer, int(weekday), hour, int(month), week, 1)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not iterate through commits: %w", err)
	}

	return activity, nil
}

// GetRepoName extracts the repository name from its path
func GetRepoName(repoPath string) string {
	base := filepath.Base(repoPath)         // Get the last component of the path
	return strings.TrimSuffix(base, ".git") // Remove ".git" suffix if present
}

func GetMultiRepoName(repoPaths []string) string {
	names := []string{}
	for _, path := range repoPaths {
		names = append(names, GetRepoName(path))
	}
	return strings.Join(names, "_and_")
}

func AnalyzeLinesInRange(repoPath string, start, end time.Time, aliases DeveloperAliases) (*CommitActivity, error) {
	activity := NewCommitActivity()

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("could not open repository: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("could not get repository head: %w", err)
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve commits: %w", err)
	}

	err = commitIter.ForEach(func(c *object.Commit) error {
		commitTime := c.Author.When
		authorEmail := strings.ToLower(c.Author.Email)

		// Map alias to developer name
		developer, exists := aliases[authorEmail]
		if !exists {
			developer = "Unknown"
		}

		// Filter commits based on date range
		if (!start.IsZero() && commitTime.Before(start)) || (!end.IsZero() && commitTime.After(end)) {
			return nil // Skip this commit
		}

		// Get the diff stats for the commit
		stats, err := c.Stats()
		if err != nil {
			return fmt.Errorf("could not get diff stats: %w", err)
		}

		// Aggregate added and deleted lines
		added, deleted := 0, 0
		for _, stat := range stats {
			added += stat.Addition
			deleted += stat.Deletion
		}

		// Increment activity data per developer
		weekday := int(commitTime.Weekday())
		hour := commitTime.Hour()
		month := int(commitTime.Month()) - 1 // Month() is 1-based, adjust to 0-based
		_, week := commitTime.ISOWeek()

		// Add activity
		if weekday >= 0 && weekday < 7 && hour >= 0 && hour < 24 && month >= 0 && month < 12 && week >= 0 && week < 53 {
			activity.AddActivity(developer, weekday, hour, month, week, added+deleted)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not iterate through commits: %w", err)
	}

	return activity, nil
}

func AnalyzeRepositories(repoPaths []string, start, end time.Time, mode string, aliases DeveloperAliases) (string, *CombinedCommitActivity) {
	fmt.Printf("Analyzing %d repositories in '%s' mode...\n", len(repoPaths), mode)
	combinedActivity := &CombinedCommitActivity{}

	for _, repoPath := range repoPaths {
		fmt.Printf("Analyzing repository: %s\n", repoPath)
		repoName := GetRepoName(repoPath)

		var activity *CommitActivity
		var err error

		// Choose analysis method based on mode
		if mode == "commits" {
			activity, err = AnalyzeCommitsInRange(repoPath, start, end, aliases)
		} else if mode == "lines" {
			activity, err = AnalyzeLinesInRange(repoPath, start, end, aliases)
		}

		if err != nil {
			log.Fatalf("Error analyzing %s for %s: %v", mode, repoPath, err)
		}

		combinedActivity.Add(repoName, activity)
	}

	outputPrefix := GetMultiRepoName(repoPaths)
	if outputPrefix == "" || len(outputPrefix) > 128 {
		outputPrefix = "combined"
	}

	return outputPrefix, combinedActivity
}
