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
				
				// 格式化显示为1天3小时5分15秒的形式
				let timeLeft = '';
				if (days > 0) {
					timeLeft += days + '天';
				}
				timeLeft += hours + '小时';
				timeLeft += minutes + '分';
				timeLeft += seconds + '秒';
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

// 全局变量保存可用语音列表
let availableVoices = [];

// 预加载语音列表
function loadVoices() {
	if ('speechSynthesis' in window) {
		// 获取语音列表
		availableVoices = speechSynthesis.getVoices();
		console.log('预加载语音列表，可用语音数量:', availableVoices.length);
		
		// 等待voiceschanged事件确保语音加载完成
		speechSynthesis.onvoiceschanged = () => {
			availableVoices = speechSynthesis.getVoices();
			console.log('语音列表更新，可用语音数量:', availableVoices.length);
		};
	}
}

// 简单的朗读函数
function speakText(text, onStart, onEnd, onError) {
	if (!('speechSynthesis' in window)) {
		console.error('浏览器不支持Web Speech API');
		if (onError) onError(new Error('浏览器不支持Web Speech API'));
		return;
	}

	// 取消任何正在进行的朗读
	speechSynthesis.cancel();

	// 创建SpeechSynthesisUtterance实例
	const utterance = new SpeechSynthesisUtterance(text);

	// 设置基本属性
	utterance.lang = 'zh-CN';
	utterance.rate = 0.9; // 稍慢一点以便更清晰
	utterance.volume = 1.0;
	utterance.pitch = 1.0;

	// 添加事件监听器
	if (onStart) {
		utterance.onstart = onStart;
	}
	if (onEnd) {
		utterance.onend = onEnd;
	}
	if (onError) {
		utterance.onerror = onError;
	}

	// 尝试设置语音
	const chineseVoice = availableVoices.find(voice => 
		voice.lang.includes('zh') || 
		voice.name.includes('Chinese') || 
		voice.name.includes('中文') ||
		voice.localService
	);

	if (chineseVoice) {
		utterance.voice = chineseVoice;
		console.log('使用语音:', chineseVoice.name, chineseVoice.lang);
	} else if (availableVoices.length > 0) {
		// 使用第一个可用语音
		utterance.voice = availableVoices[0];
		console.log('使用默认语音:', availableVoices[0].name, availableVoices[0].lang);
	}

	// 开始朗读
	console.log('开始朗读文本:', text);
	speechSynthesis.speak(utterance);
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

	// 预加载语音列表
	loadVoices();

	// 实现Web Speech API朗读功能
	const readAloudButtons = document.querySelectorAll('.read-aloud-btn');
	readAloudButtons.forEach(button => {
		button.addEventListener('click', function() {
			console.log('朗读按钮被点击');
			// 获取要朗读的文本
			const text = this.getAttribute('data-text');
			console.log('要朗读的文本:', text);
			
			if (text) {
				const originalText = button.textContent;
				button.textContent = '⏳'; // 显示加载状态
				
				// 使用封装的朗读函数
				speakText(
					text,
					() => {
						console.log('朗读开始');
						button.textContent = '🔊';
					},
					() => {
						console.log('朗读完成');
						button.textContent = '🔊';
					},
					(error) => {
						console.error('朗读出错:', error);
						button.textContent = '🔊';
						alert('朗读时出错: ' + (error.message || error));
					}
				);
			}
		});
	});
});