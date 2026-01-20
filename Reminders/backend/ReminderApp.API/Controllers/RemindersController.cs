using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using ReminderApp.API.Data;
using ReminderApp.API.Models;

namespace ReminderApp.API.Controllers
{
    [ApiController]
    [Route("api/[controller]")]
    public class RemindersController : ControllerBase
    {
        private readonly AppDbContext _context;

        public RemindersController(AppDbContext context)
        {
            _context = context;
        }

        [HttpGet]
        public async Task<ActionResult<IEnumerable<Reminder>>> GetReminders()
        {
            return await _context.Reminders.Include(r => r.Assignees).ToListAsync();
        }

        [HttpPost]
        public async Task<ActionResult<Reminder>> PostReminder(Reminder reminder)
        {
            // Attach existing users to avoid duplicates if they come in as simple objects
            // Simple logic: if IDs are present (even though valid JSON for collection is complex), 
            // for now, we assume the frontend sends a format EF can handle or we fix it later.
            // Better approach: Look up by ID if the User object only has ID.
            
           if (reminder.Assignees != null && reminder.Assignees.Any())
           {
               var userIds = reminder.Assignees.Select(u => u.Id).ToList();
               var existingUsers = await _context.Users.Where(u => userIds.Contains(u.Id)).ToListAsync();
               reminder.Assignees = existingUsers;
           }

            _context.Reminders.Add(reminder);
            await _context.SaveChangesAsync();

            return CreatedAtAction(nameof(GetReminders), new { id = reminder.Id }, reminder);
        }

        [HttpDelete("{id}")]
        public async Task<IActionResult> DeleteReminder(int id)
        {
            var reminder = await _context.Reminders.FindAsync(id);
            if (reminder == null) return NotFound();

            _context.Reminders.Remove(reminder);
            await _context.SaveChangesAsync();
            return NoContent();
        }
    }
}
