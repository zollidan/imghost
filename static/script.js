function updateCountdown() {
    document.querySelectorAll('#fileTable tr[data-expires]').forEach(row => {
        const expires = parseInt(row.getAttribute('data-expires')) * 1000;
        const diff = expires - Date.now();
        const cell = row.querySelector('.countdown');
        if (diff <= 0) {
            cell.textContent = 'expired';
        } else {
            let sec = Math.floor(diff / 1000);
            const min = Math.floor(sec / 60);
            sec = sec % 60;
            cell.textContent = `${min}:${sec.toString().padStart(2, '0')}`;
        }
    });
}
setInterval(updateCountdown, 1000);
updateCountdown();
