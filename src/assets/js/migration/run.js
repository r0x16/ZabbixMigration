document.addEventListener("DOMContentLoaded", function() {
    let textarea = document.getElementById('runLogArea');
    let scrollHeight = textarea.scrollHeight;
    let animationDuration = 200; // milliseconds

    // Calculate the number of frames for the animation
    let frames = Math.ceil(animationDuration / 16.7); // Assuming 60 frames per second

    // Calculate the distance to scroll per frame
    let distancePerFrame = scrollHeight / frames;

    // Define the animation function
    function animateScroll() {
        if (textarea.scrollTop < scrollHeight) {
            textarea.scrollTop += distancePerFrame;
            requestAnimationFrame(animateScroll);
        }
    }

    // Start the animation
    requestAnimationFrame(animateScroll);
});

