// 主JavaScript文件

// 计算并显示任务倒计时
function updateTaskCountdowns() {
	// 找到所有任务过期时间元素
	const expiryElements = document.querySelectorAll('.task-expiry');
	expiryElements.forEach((element) => {
		const expiryTimeStr = element.getAttribute('data-expiry');
		if (expiryTimeStr) {
			// 解析截止时间 - 添加更健壮的日期解析
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

// 主JavaScript文件

// 页面加载完成后执行
window.addEventListener('DOMContentLoaded', function() {
	console.log('页面加载完成，开始初始化脚本...');
	
	// 初始化任务倒计时
	updateTaskCountdowns();
	// 每秒更新一次倒计时
	setInterval(updateTaskCountdowns, 1000);
	
	// 为所有删除按钮添加确认提示
	const deleteButtons = document.querySelectorAll('.delete-btn');
	deleteButtons.forEach(button => {
		button.addEventListener('click', function(e) {
			if (!confirm('确定要删除这个任务吗？')) {
				e.preventDefault();
			}
		});
	});

	// 为所有表单添加CSRF保护（简化版）
	const forms = document.querySelectorAll('form');
	forms.forEach(form => {
		// 在实际应用中应该添加CSRF令牌
		form.addEventListener('submit', function() {
			// 可以在这里添加加载状态
			console.log('表单提交中...');
		});
	});

	// 简单测试函数 - 用于直接测试模态框显示
	window.testModal = function() {
		console.log('测试函数被调用');
		const modal = document.getElementById('createTaskModal');
		if (modal) {
			console.log('找到模态框，尝试显示');
			modal.style.display = 'block';
			console.log('模态框样式:', modal.style);
		} else {
			console.log('未找到模态框');
		}
	};
	
	console.log('测试函数已注册，请在控制台输入 testModal() 测试');
	
	// 直接通过ID获取创建任务按钮并绑定事件
	const createTaskBtn = document.getElementById('create-task-button');
	const createTaskModal = document.getElementById('createTaskModal');
	const closeModalButtons = document.querySelectorAll('.close-modal');
	const cancelBtn = document.querySelector('.cancel-btn');
	const closeX = document.querySelector('.close-modal');
	
	console.log('按钮元素 (by ID):', createTaskBtn);
	console.log('按钮元素 (by class):', document.querySelector('.create-task-btn'));
	console.log('模态框元素:', createTaskModal);
	console.log('关闭按钮数量:', closeModalButtons.length);
	
	// 如果通过ID没找到，尝试通过class查找
	if (!createTaskBtn) {
		console.log('通过ID未找到按钮，尝试通过class查找...');
		const buttons = document.querySelectorAll('.create-task-btn');
		console.log('通过class找到的按钮数量:', buttons.length);
		
		// 为所有找到的按钮绑定事件
		buttons.forEach((btn, index) => {
			console.log('为按钮', index, '绑定事件');
			btn.addEventListener('click', function(e) {
				e.preventDefault();
				console.log('点击了创建任务按钮（class），尝试打开模态框');
				if (createTaskModal) {
					createTaskModal.style.display = 'block';
					console.log('模态框显示状态:', createTaskModal.style.display);
				}
			});
		});
	} else {
		// 打开模态框
		createTaskBtn.addEventListener('click', function(e) {
			e.preventDefault();
			console.log('点击了创建任务按钮（ID），尝试打开模态框');
			if (createTaskModal) {
				createTaskModal.style.display = 'block';
				console.log('模态框显示状态:', createTaskModal.style.display);
			}
		});
	}
	
	// 关闭模态框 - 为所有关闭按钮添加事件
	if (closeModalButtons.length > 0) {
		closeModalButtons.forEach(button => {
			button.addEventListener('click', function(e) {
				e.preventDefault();
				console.log('点击了关闭按钮');
				if (createTaskModal) {
					createTaskModal.style.display = 'none';
				}
			});
		});
	}
	
	// 单独为取消按钮添加事件
	if (cancelBtn) {
		cancelBtn.addEventListener('click', function(e) {
			e.preventDefault();
			console.log('点击了取消按钮');
			if (createTaskModal) {
				createTaskModal.style.display = 'none';
			}
		});
	}
	
	// 单独为X按钮添加事件
	if (closeX) {
		closeX.addEventListener('click', function(e) {
			e.preventDefault();
			console.log('点击了X按钮');
			if (createTaskModal) {
				createTaskModal.style.display = 'none';
			}
		});
	}
	
	// 点击模态框外部关闭
	window.addEventListener('click', function(event) {
		if (event.target === createTaskModal) {
			console.log('点击了模态框外部');
			createTaskModal.style.display = 'none';
		}
	});

	// 为所有导航链接添加平滑滚动效果
	const navLinks = document.querySelectorAll('.nav-link');
	navLinks.forEach(link => {
		link.addEventListener('click', function(e) {
			const href = this.getAttribute('href');
			// 只有内部链接才应用平滑滚动
			if (href.startsWith('#')) {
				e.preventDefault();
				const targetId = href.substring(1);
				const targetElement = document.getElementById(targetId);
				if (targetElement) {
					targetElement.scrollIntoView({ behavior: 'smooth' });
				}
			}
		});
	});

	// 添加简单的动画效果
	addHoverEffects();

	// 更新年份
	updateYear();
});

// 添加悬停效果
function addHoverEffects() {
	const cards = document.querySelectorAll('.task-card, .item-card, .feature-card');
	cards.forEach(card => {
		card.addEventListener('mouseenter', function() {
			this.style.transform = 'translateY(-5px)';
			this.style.boxShadow = '0 10px 20px rgba(0, 0, 0, 0.5)';
		});
		card.addEventListener('mouseleave', function() {
			this.style.transform = 'translateY(0)';
			this.style.boxShadow = 'none';
		});
	});
}

// 更新页脚年份
function updateYear() {
	const yearElements = document.querySelectorAll('.minecraft-footer p');
	const currentYear = new Date().getFullYear();
	yearElements.forEach(element => {
		if (element.textContent.includes('{{.Year}}')) {
		element.textContent = element.textContent.replace('{{.Year}}', currentYear);
		}
	});
}

// 任务完成提交函数
function submitTaskCompletion(taskId) {
	// 在实际应用中应该使用AJAX提交
	console.log('提交任务完成:', taskId);
	// 这里可以添加加载动画
	setTimeout(() => {
		// 模拟提交成功
		alert('任务提交成功，请等待家长确认！');
		window.location.reload();
	}, 1000);
}

// 物品兑换函数
function exchangeItem(itemId) {
	// 在实际应用中应该使用AJAX提交
	console.log('兑换物品:', itemId);
	// 这里可以添加加载动画
}

// 简单的错误处理函数
function handleError(error) {
	console.error('发生错误:', error);
	alert('操作失败，请稍后重试！');
}