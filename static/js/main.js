// 主JavaScript文件

// 计算并显示任务倒计时
function updateTaskCountdowns() {
	// 找到所有任务过期时间元素
	const expiryElements = document.querySelectorAll('.task-expiry');
	expiryElements.forEach((element) => {
		const expiryTimeStr = element.getAttribute('data-expiry');
		if (expiryTimeStr) {
			// 解析截止时间
			let expiryTime;
			try {
				expiryTime = new Date(expiryTimeStr);
			} catch (error) {
				console.error('解析日期失败:', error);
				return;
			}
			
			const now = new Date();
			const diff = expiryTime - now;
			
			// 如果已经过期，显示已过期
			if (diff <= 0) {
				element.textContent = '已过期';
				element.style.color = 'red';
			} else {
				// 计算天、时、分、秒
				const days = Math.floor(diff / (1000 * 60 * 60 * 24));
				const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
				const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
				const seconds = Math.floor((diff % (1000 * 60)) / 1000);
				
				// 格式化显示，确保所有时间单位都显示两位数
				let timeLeft = '';
				if (days > 0) {
					timeLeft += days + '天';
				}
				timeLeft += hours.toString().padStart(2, '0') + '时';
				timeLeft += minutes.toString().padStart(2, '0') + '分';
				timeLeft += seconds.toString().padStart(2, '0') + '秒';
				element.textContent = '剩余: ' + timeLeft;
			}
		}
	});
}

// 格式化任务开始时间
function formatTaskStartTimes() {
	// 找到所有任务开始时间元素
	const startElements = document.querySelectorAll('.task-start');
	startElements.forEach((element) => {
		const startTimeStr = element.getAttribute('data-start');
		if (startTimeStr) {
			// 解析开始时间
			let startTime;
			try {
				startTime = new Date(startTimeStr);
			} catch (error) {
				console.error('解析日期失败:', error);
				return;
			}
			
			// 格式化开始时间为YYYY-MM-DD HH:MM:SS格式
			const year = startTime.getFullYear();
			const month = String(startTime.getMonth() + 1).padStart(2, '0');
			const day = String(startTime.getDate()).padStart(2, '0');
			const hours = String(startTime.getHours()).padStart(2, '0');
			const minutes = String(startTime.getMinutes()).padStart(2, '0');
			const seconds = String(startTime.getSeconds()).padStart(2, '0');
			
			const formattedTime = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
			element.textContent = '开始时间: ' + formattedTime;
		}
	});
}

// 页面加载完成后执行
window.addEventListener('DOMContentLoaded', function() {
	// 初始化任务倒计时
	updateTaskCountdowns();
	// 每秒更新一次倒计时
	setInterval(updateTaskCountdowns, 1000);
	
	// 格式化任务开始时间
	formatTaskStartTimes();
	
	// 为所有删除按钮添加确认提示
	const deleteButtons = document.querySelectorAll('.delete-btn');
	deleteButtons.forEach(button => {
		button.addEventListener('click', function(e) {
			if (!confirm('确定要删除这个任务吗？')) {
				e.preventDefault();
			}
		});
	});

	// 为所有表单添加CSRF保护
	const forms = document.querySelectorAll('form');
	forms.forEach(form => {
		// 在实际应用中应该添加CSRF令牌
		form.addEventListener('submit', function() {
			// 可以在这里添加加载状态
		});
	});
});