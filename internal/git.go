package internal

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

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
	Weekdays [7]int
	Hours    [24]int
	Months   [12]int
	Weeks    [53]int
}

func (ca *CommitActivity) Combine(other *CommitActivity) {
	for i := 0; i < len(ca.Weekdays); i++ {
		ca.Weekdays[i] += other.Weekdays[i]
	}
	for i := 0; i < len(ca.Hours); i++ {
		ca.Hours[i] += other.Hours[i]
	}
	for i := 0; i < len(ca.Months); i++ {
		ca.Months[i] += other.Months[i]
	}
	for i := 0; i < len(ca.Weeks); i++ {
		ca.Weeks[i] += other.Weeks[i]
	}
}

func AnalyzeCommits(repoPath string) (*CommitActivity, error) {
	activity := &CommitActivity{}

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

		// Weekday and hour
		weekday := commitTime.Weekday()
		hour := commitTime.Hour()

		// Month (0 = January, 11 = December)
		month := commitTime.Month() - 1 // `time.Month` is 1-based

		// Week number (0 = first week of year)
		_, week := commitTime.ISOWeek()

		activity.Weekdays[int(weekday)]++
		activity.Hours[hour]++
		activity.Months[int(month)]++
		activity.Weeks[week]++

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not iterate through commits: %w", err)
	}

	return activity, nil
}

func AnalyzeCommitsInRange(repoPath string, start, end time.Time) (*CommitActivity, error) {
	activity := &CommitActivity{}

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

		// Filter commits based on date range
		if (!start.IsZero() && commitTime.Before(start)) || (!end.IsZero() && commitTime.After(end)) {
			return nil // Skip this commit
		}

		// Process the commit
		weekday := commitTime.Weekday()
		hour := commitTime.Hour()
		month := commitTime.Month() - 1
		_, week := commitTime.ISOWeek()

		activity.Weekdays[int(weekday)]++
		activity.Hours[hour]++
		activity.Months[int(month)]++
		activity.Weeks[week]++

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not iterate through commits: %w", err)
	}

	return activity, nil
}

// CreateBarChart creates a bar chart from the given data
func CreateBarChart(values []int, labels []string, title, filename string) error {
	pts := make(plotter.Values, len(values))
	for i, v := range values {
		pts[i] = float64(v)
	}

	p := plot.New()
	p.Title.Text = title
	bar, err := plotter.NewBarChart(pts, vg.Points(20))
	if err != nil {
		return fmt.Errorf("could not create bar chart: %w", err)
	}

	bar.LineStyle.Width = vg.Length(0)
	bar.Color = colorPalette[1%len(colorPalette)]

	p.Add(bar)
	p.NominalX(labels...)

	// Save the chart to the specified file format (determined by file extension)
	err = p.Save(10*vg.Inch, 4*vg.Inch, filename)
	if err != nil {
		return fmt.Errorf("could not save chart: %w", err)
	}

	return nil
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

type ModificationCache struct {
	timestamps map[string]time.Time
}

func NewModificationCache() *ModificationCache {
	return &ModificationCache{
		timestamps: make(map[string]time.Time),
	}
}

func (mc *ModificationCache) Get(fileName string, commitHash plumbing.Hash) (time.Time, bool) {
	key := fmt.Sprintf("%s:%s", commitHash.String(), fileName)
	t, exists := mc.timestamps[key]
	return t, exists
}

func (mc *ModificationCache) Set(fileName string, commitHash plumbing.Hash, timestamp time.Time) {
	key := fmt.Sprintf("%s:%s", commitHash.String(), fileName)
	mc.timestamps[key] = timestamp
}

func AnalyzeLinesInRange(repoPath string, start, end time.Time) (*CommitActivity, error) {
	activity := &CommitActivity{}
	cache := NewModificationCache()

	// Open the Git repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("could not open repository: %w", err)
	}

	// Get the repository head
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("could not get repository head: %w", err)
	}

	// Get the commit history
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve commits: %w", err)
	}

	// Iterate through the commits
	err = commitIter.ForEach(func(c *object.Commit) error {
		// skip commit if before start
		if !start.IsZero() && c.Author.When.Before(start) {
			return nil
		}

		// get files in commit
		files, err := c.Files()
		if err != nil {
			return fmt.Errorf("could not get files for commit: %w", err)
		}

		// iterate through files
		err = files.ForEach(func(file *object.File) error {
			// Check cache first
			lastModTime, cached := cache.Get(file.Name, c.Hash)
			if !cached {
				lastModTime, err = getLastModificationTime(repo, file.Name, c.Hash)
				if err != nil {
					return fmt.Errorf("could not get last modification time for file %s: %w", file.Name, err)
				}
				cache.Set(file.Name, c.Hash, lastModTime)
			}

			// Filter by the provided date range
			if (!start.IsZero() && lastModTime.Before(start)) || (!end.IsZero() && lastModTime.After(end)) {
				return nil // Skip this file
			}

			// Get diff stats for this commit
			stats, err := c.Stats()
			if err != nil {
				return fmt.Errorf("could not get diff stats: %w", err)
			}

			// Aggregate added and deleted lines for the file
			added, deleted := 0, 0
			for _, stat := range stats {
				if stat.Name == file.Name {
					added += stat.Addition
					deleted += stat.Deletion
				}
			}

			// Increment activity data
			weekday := lastModTime.Weekday()
			hour := lastModTime.Hour()
			month := lastModTime.Month() - 1
			_, week := lastModTime.ISOWeek()

			activity.Weekdays[int(weekday)] += added + deleted
			activity.Hours[hour] += added + deleted
			activity.Months[int(month)] += added + deleted
			activity.Weeks[week] += added + deleted

			return nil
		})

		if err != nil {
			return fmt.Errorf("error processing files in commit: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not iterate through commits: %w", err)
	}

	return activity, nil
}

func getLastModificationTime(repo *git.Repository, fileName string, currentCommit plumbing.Hash) (time.Time, error) {
	// Retrieve the commit log for the file
	logIter, err := repo.Log(&git.LogOptions{
		FileName: &fileName,
		From:     currentCommit,
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("could not retrieve log for file %s: %w", fileName, err)
	}

	// Iterate through the log to find the latest commit for this file
	var lastModTime time.Time
	err = logIter.ForEach(func(commit *object.Commit) error {
		lastModTime = commit.Author.When
		// Stop after the first (latest) commit
		return storer.ErrStop
	})

	if err != nil && err != storer.ErrStop {
		return time.Time{}, fmt.Errorf("could not iterate through log for file %s: %w", fileName, err)
	}

	return lastModTime, nil
}

func AnalyzeRepositories(repoPaths []string, start, end time.Time, mode string) (string, *CombinedCommitActivity) {
	fmt.Printf("Analyzing %d repositories in '%s' mode...\n", len(repoPaths), mode)
	combinedActivity := &CombinedCommitActivity{}

	for _, repoPath := range repoPaths {
		fmt.Printf("Analyzing repository: %s\n", repoPath)
		repoName := GetRepoName(repoPath)

		var activity *CommitActivity
		var err error

		// Choose analysis method based on mode
		if mode == "commits" {
			activity, err = AnalyzeCommitsInRange(repoPath, start, end)
		} else if mode == "lines" {
			activity, err = AnalyzeLinesInRange(repoPath, start, end)
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
