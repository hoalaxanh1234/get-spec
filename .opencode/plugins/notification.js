export const NotificationPlugin = async ({ $ }) => {
  return {
    event: async ({ event }) => {
      // Trigger when the AI finishes its work
      if (event.type === "session.idle") {
        const title = "OpenCode AI";
        const message = "Task completed successfully!";

        // We add 'Import-Module' and 'ExecutionPolicy Bypass' to ensure it works from WSL
        await $`powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "Import-Module BurntToast; New-BurntToastNotification -Text '${title}', '${message}'"`.catch((e) => {
          console.error("Notification failed:", e);
        });
      }
    },
  };
};