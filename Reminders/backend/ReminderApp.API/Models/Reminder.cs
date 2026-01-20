namespace ReminderApp.API.Models
{
    public class Reminder
    {
        public int Id { get; set; }
        public required string Description { get; set; }
        public DateTime TargetDate { get; set; }
        public bool IsRecurring { get; set; }
        public string? RecurrencePattern { get; set; } // i.e., "WEEKLY;INTERVAL=1;BYDAY=MO" or JSON configuration
        
        // Simplified foreign key relationship for SQLite (comma separated IDs or simple join table later if needed)
        // For simplicity in this demo, we'll store assigned user IDs as a json list or separate collection
        // But EF Core handles lists well if configured. Let's do a proper relationship?
        // User requested "assigned to one or more persons".
        public List<User> Assignees { get; set; } = new();
    }
}
