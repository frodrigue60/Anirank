const axios = require('axios');

async function checkActivities() {
    try {
        const response = await axios.get('http://localhost:8080/api/activities/recent');
        const activities = response.data.data;
        
        if (!activities || activities.length === 0) {
            console.log("No activities found.");
            return;
        }

        activities.slice(0, 3).forEach((activity, index) => {
            console.log(`Activity ${index + 1}:`);
            console.log(`  User: ${activity.user?.name} (Avatar: ${activity.user?.avatar_url})`);
            console.log(`  Target: ${activity.target_type}`);
            if (activity.song) {
                console.log(`  Song: ${activity.song.name} (Anime Cover: ${activity.song.anime?.cover_url})`);
            }
            if (activity.artist) {
                console.log(`  Artist: ${activity.artist.name} (Avatar: ${activity.artist.avatar_url})`);
            }
            console.log('---');
        });
    } catch (error) {
        console.error("Error fetching activities:", error.message);
    }
}

checkActivities();
