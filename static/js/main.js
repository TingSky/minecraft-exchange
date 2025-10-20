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
let voicesLoaded = false;

// 预加载语音列表 - 优化版本
function loadVoices() {
	if ('speechSynthesis' in window) {
		// 首次尝试获取语音列表
		updateVoicesList();
		
		// 等待voiceschanged事件确保语音加载完成
		speechSynthesis.onvoiceschanged = () => {
			updateVoicesList();
		};
		
		// 额外的超时重试机制，确保在iOS上也能获取到语音
		setTimeout(() => {
			if (!voicesLoaded) {
				console.log('尝试重新加载语音列表...');
				updateVoicesList();
			}
		}, 1000);
	}
}

// 更新语音列表的辅助函数
function updateVoicesList() {
	try {
		const voices = speechSynthesis.getVoices();
		availableVoices = voices;
		voicesLoaded = voices.length > 0;
		console.log('语音列表更新，可用语音数量:', voices.length);
		// 打印所有可用语音信息，方便调试
		voices.forEach((voice, index) => {
			console.log(`语音 ${index}:`, voice.name, voice.lang, voice.localService);
		});
	} catch (error) {
		console.error('获取语音列表时出错:', error);
	}
}

// 优化的朗读函数，兼容iPad/iOS
function speakText(text, onStart, onEnd, onError) {
	if (!('speechSynthesis' in window)) {
		const error = new Error('浏览器不支持Web Speech API');
		console.error(error.message);
		if (onError) onError(error);
		return;
	}

	try {
		// 取消任何正在进行的朗读
		speechSynthesis.cancel();

		// 重新获取语音列表（iOS可能需要在每次使用前重新获取）
		updateVoicesList();

		// 创建SpeechSynthesisUtterance实例
		const utterance = new SpeechSynthesisUtterance(text);

		// 设置基本属性 - 简化设置以提高兼容性
		utterance.lang = 'zh-CN';
		utterance.rate = 0.9;
		utterance.volume = 1.0;
		// 在iOS上不设置pitch，避免兼容性问题

		// 添加事件监听器
		utterance.onstart = function(event) {
			console.log('朗读开始事件触发');
			if (onStart) onStart(event);
		};
		utterance.onend = function(event) {
			console.log('朗读结束事件触发');
			if (onEnd) onEnd(event);
		};
		utterance.onerror = function(event) {
			console.error('朗读错误事件触发:', event.error);
			if (onError) onError(new Error(event.error));
		};
		utterance.onpause = function() {
			console.log('朗读暂停');
		};
		utterance.onresume = function() {
			console.log('朗读恢复');
		};

		// 优化语音选择逻辑，更好地适应iOS
		let selectedVoice = null;
		
		// 1. 尝试选择中文语音
		if (availableVoices.length > 0) {
			// 优先选择本地服务语音（在iOS上可能更可靠）
			selectedVoice = availableVoices.find(voice => 
				voice.localService && 
				(voice.lang === 'zh-CN' || voice.lang === 'zh-Hans-CN')
			);
			
			// 如果没找到本地中文语音，尝试其他中文语音
			if (!selectedVoice) {
				selectedVoice = availableVoices.find(voice => 
					voice.lang.includes('zh') || 
					voice.name.includes('Chinese') || 
					voice.name.includes('中文')
				);
			}
			
			// 如果还没找到，使用第一个可用语音
			if (!selectedVoice) {
				selectedVoice = availableVoices[0];
			}
			
			// 设置语音
			try {
				utterance.voice = selectedVoice;
				console.log('使用语音:', selectedVoice.name, selectedVoice.lang, selectedVoice.localService);
			} catch (voiceError) {
				console.warn('设置语音失败:', voiceError);
				// 即使设置语音失败，也继续尝试朗读
			}
		}

		// 开始朗读 - 使用try-catch确保在任何情况下都不会崩溃
		console.log('开始朗读文本:', text);
		speechSynthesis.speak(utterance);
		
		// iOS特殊处理：如果语音没有立即播放，尝试一个简短的延迟后再次触发
		const checkIfSpeaking = setTimeout(() => {
			if (!speechSynthesis.speaking) {
				console.log('检测到未开始朗读，尝试再次触发...');
				// 再次尝试，但这次不设置voice属性，使用系统默认
				speechSynthesis.cancel();
				const fallbackUtterance = new SpeechSynthesisUtterance(text);
				fallbackUtterance.lang = 'zh-CN';
				fallbackUtterance.rate = 0.9;
				fallbackUtterance.volume = 1.0;
				speechSynthesis.speak(fallbackUtterance);
			}
		}, 500);
		
		// 清除定时器
		utterance.onend = function(event) {
			clearTimeout(checkIfSpeaking);
			if (onEnd) onEnd(event);
		};
		
	} catch (error) {
		console.error('朗读过程中发生错误:', error);
		if (onError) onError(error);
	}
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