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

		// 简化设置，提高兼容性
		utterance.lang = 'zh-CN';
		utterance.rate = 0.9;
		utterance.volume = 1.0;

		// 添加事件监听器
		utterance.onstart = function(event) {
			console.log('朗读开始事件触发');
			if (onStart) onStart(event);
		};
		utterance.onend = function(event) {
			console.log('朗读结束事件触发');
			clearTimeout(checkIfSpeaking);
			if (onEnd) onEnd(event);
		};
		utterance.onerror = function(event) {
			console.error('朗读错误事件触发:', event.error);
			clearTimeout(checkIfSpeaking);
			if (onError) onError(new Error(event.error));
		};

		// 简单的语音选择逻辑，避免过于复杂的判断
		// 在iPad/iOS上，我们先尝试使用系统默认语音，如果不行再尝试其他方法
		console.log('开始朗读文本:', text);
		
		// 尝试直接朗读，使用默认设置
		speechSynthesis.speak(utterance);
		
		// iOS特殊处理：监控朗读状态并提供后备方案
		const checkIfSpeaking = setTimeout(() => {
			if (!speechSynthesis.speaking) {
				console.log('检测到未开始朗读，尝试使用后备方案...');
				// 取消当前朗读
				speechSynthesis.cancel();
				
				// 后备方案1: 不设置lang，让系统自动选择
				const fallbackUtterance1 = new SpeechSynthesisUtterance(text);
				fallbackUtterance1.rate = 0.9;
				fallbackUtterance1.volume = 1.0;
				
				// 添加错误处理到后备方案
				fallbackUtterance1.onerror = function(event) {
					console.error('后备方案1朗读错误:', event.error);
					clearTimeout(checkFallback);
					if (onError) onError(new Error('所有朗读方案失败'));
				};
				
				fallbackUtterance1.onend = function(event) {
					clearTimeout(checkFallback);
					if (onEnd) onEnd(event);
				};
				
				speechSynthesis.speak(fallbackUtterance1);
				
				// 再次检查后备方案是否成功
				const checkFallback = setTimeout(() => {
					if (!speechSynthesis.speaking) {
						console.log('后备方案1也失败，尝试最后方案...');
						speechSynthesis.cancel();
						
						// 最后方案: 分段朗读，避免长文本问题
						const fallbackUtterance2 = new SpeechSynthesisUtterance(text.substring(0, 200));
						fallbackUtterance2.rate = 0.9;
						fallbackUtterance2.volume = 1.0;
						
						fallbackUtterance2.onerror = function() {
							if (onError) onError(new Error('所有朗读方案失败'));
						};
						
						fallbackUtterance2.onend = function() {
							if (onEnd) onEnd();
						};
						
						speechSynthesis.speak(fallbackUtterance2);
					}
				}, 500);
			}
		}, 500);
		
	} catch (error) {
		console.error('朗读过程中发生错误:', error);
		if (onError) onError(error);
	}
}

// 局部刷新指定模块
function refreshModule(moduleSelector) {
    const module = document.querySelector(moduleSelector);
    if (module) {
        // 添加刷新动画效果
        module.style.opacity = '0.5';
        module.style.transition = 'opacity 0.3s ease';
        
        // 模拟局部刷新（实际项目中可以根据需要实现更复杂的局部刷新逻辑）
        setTimeout(() => {
            module.style.opacity = '1';
            // 这里可以根据需要添加更具体的局部刷新逻辑
            // 例如，重新加载特定区域的内容
            window.location.reload(); // 暂时保持全局刷新，后续可以优化为真正的局部刷新
        }, 300);
    }
}

// 通用AJAX表单提交函数
function ajaxFormSubmit(form, successCallback, errorCallback) {
    form.addEventListener('submit', function(e) {
        // 根据表单action确定confirm消息
        const action = form.action;
        let confirmMessage;
        
        if (action.includes('/exchange')) {
            confirmMessage = '确认要兑换此物品吗？';
        } else if (action.includes('/delete_task')) {
            confirmMessage = '确定要删除这个任务吗？';
        } else if (action.includes('/delete_task_template')) {
            confirmMessage = '确定要删除这个模板吗？';
        } else if (action.includes('/delete_item')) {
            confirmMessage = '确定要删除这个物品吗？';
        }
        
        // 显示confirm对话框，如果用户取消，则阻止表单提交
        if (confirmMessage && !confirm(confirmMessage)) {
            e.preventDefault();
            return;
        }
        
        e.preventDefault();
        
        // 显示加载状态
        const submitButton = form.querySelector('button[type="submit"]');
        const originalButtonText = submitButton ? submitButton.textContent : '';
        if (submitButton) {
            submitButton.disabled = true;
            submitButton.textContent = '处理中...';
        }
        
        const formData = new FormData(form);
        const url = form.action;
        
        fetch(url, {
            method: 'POST',
            body: formData,
            credentials: 'include'
        })
        .then(response => {
            // 首先检查HTTP状态码
            if (response.ok) {
                // 成功响应，尝试解析JSON
                return response.json().then(data => ({
                    success: true,
                    data: data
                })).catch(() => ({
                    // JSON解析失败，但HTTP状态码是200，视为成功
                    success: true,
                    data: {
                        success: true,
                        message: '操作成功',
                        refresh: true
                    }
                }));
            } else {
                // 失败响应，尝试解析JSON获取错误信息
                return response.json().then(data => ({
                    success: false,
                    data: data
                })).catch(() => ({
                    // JSON解析失败，使用HTTP状态码作为错误信息
                    success: false,
                    data: {
                        success: false,
                        message: `HTTP错误 ${response.status}`
                    }
                }));
            }
        })
        .then(result => {
            // 恢复按钮状态
            if (submitButton) {
                submitButton.disabled = false;
                submitButton.textContent = originalButtonText;
            }
            
            const data = result.data;
            
            if (result.success && (data.success || data.redirect || data.refresh)) {
                // 检查表单是否在浮窗中
                const modal = form.closest('.modal');
                if (modal) {
                    // 关闭浮窗
                    modal.style.display = 'none';
                    // 重置表单
                    form.reset();
                }
                
                if (successCallback) successCallback(data);
                // 显示成功消息
                showMessage(data.message || '操作成功');
                // 如果有redirect字段，重定向到指定URL
                if (data.redirect) {
                    window.location.href = data.redirect;
                }
                // 如果需要刷新，根据表单类型决定刷新方式
                if (data.refresh) {
                    // 根据表单action决定刷新哪个模块
                    const action = form.action;
                    if (action.includes('/create_task') || action.includes('/verify_task') || action.includes('/delete_task')) {
                        // 刷新任务相关模块
                        refreshModule('.task-table');
                    } else if (action.includes('/create_item') || action.includes('/update_item') || action.includes('/delete_item')) {
                        // 刷新物品相关模块
                        refreshModule('.item-grid, .task-table');
                    } else if (action.includes('/exchange_reward')) {
                        // 刷新兑换记录模块
                        refreshModule('.exchange-table');
                    } else {
                        // 默认刷新整个页面
                        window.location.reload();
                    }
                }
            } else {
                if (errorCallback) {
                    errorCallback(data);
                } else {
                    showMessage(data.message || '操作失败', 'error');
                }
            }
        })
        .catch(error => {
            // 恢复按钮状态
            if (submitButton) {
                submitButton.disabled = false;
                submitButton.textContent = originalButtonText;
            }
            
            if (errorCallback) {
                errorCallback({ message: error.message });
            } else {
                showMessage('操作失败: ' + error.message, 'error');
            }
        });
    });
}

// 显示消息函数
function showMessage(message, type = 'success') {
    // 创建消息元素
    const messageEl = document.createElement('div');
    messageEl.className = `message ${type}`;
    messageEl.textContent = message;
    
    // 添加样式
    messageEl.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 15px 20px;
        border-radius: 5px;
        color: white;
        font-weight: bold;
        z-index: 10000;
        animation: slideIn 0.3s ease-out;
        box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    `;
    
    // 根据类型设置背景色
    if (type === 'success') {
        messageEl.style.backgroundColor = '#4CAF50';
    } else if (type === 'error') {
        messageEl.style.backgroundColor = '#f44336';
    } else if (type === 'warning') {
        messageEl.style.backgroundColor = '#ff9800';
    }
    
    // 添加到页面
    document.body.appendChild(messageEl);
    
    // 3秒后自动移除
    setTimeout(() => {
        messageEl.style.animation = 'slideOut 0.3s ease-in';
        setTimeout(() => {
            messageEl.remove();
        }, 300);
    }, 3000);
}

// 添加CSS动画
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
    
    @keyframes slideOut {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(100%);
            opacity: 0;
        }
    }
`;
document.head.appendChild(style);

// 加载任务数据
async function loadTasksData() {
    try {
        const response = await fetch('/tasks_data', {
            method: 'GET',
            credentials: 'include',
            headers: {
                'X-Requested-With': 'XMLHttpRequest'
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP错误 ${response.status}`);
        }
        
        const data = await response.json();
        
        if (data.success) {
            renderTasks(data.data);
        } else {
            showMessage(data.message || '加载任务数据失败', 'error');
        }
    } catch (error) {
        console.error('加载任务数据失败:', error);
        showMessage('加载任务数据失败: ' + error.message, 'error');
    }
}

// 渲染任务数据
function renderTasks(data) {
    // 渲染可用任务
    const availableTasksContainer = document.getElementById('available-tasks-container');
    if (availableTasksContainer) {
        let html = '';
        data.AvailableTasks.forEach(task => {
            html += `
                <div class="task-card">
                    <h3 class="task-title">${task.Title}</h3>
                    <p class="task-description">${task.Description}</p>
                    <div class="task-meta">
                        <span class="task-difficulty difficulty-${task.Difficulty}">${task.Difficulty === 'easy' ? '简单' : task.Difficulty === 'medium' ? '中等' : '困难'}</span>
                        <span class="task-reward">奖励: ${task.Reward} 绿宝石</span>
                    </div>
                    <div class="task-actions">
                        <form action="/claim_task" method="post">
                            <input type="hidden" name="task_id" value="${task.ID}">
                            <button type="submit" class="minecraft-btn claim-btn">领取任务</button>
                        </form>
                    </div>
                </div>
            `;
        });
        availableTasksContainer.innerHTML = html;
    }
    
    // 渲染已领取任务
    const claimedTasksContainer = document.getElementById('claimed-tasks-container');
    if (claimedTasksContainer) {
        let html = '';
        data.ClaimedTasks.forEach(task => {
            html += `
                <div class="task-card claimed">
                    <h3 class="task-title">${task.Title}</h3>
                    <p class="task-description">${task.Description}</p>
                    <div class="task-meta">
                        <span class="task-difficulty difficulty-${task.Difficulty}">${task.Difficulty === 'easy' ? '简单' : task.Difficulty === 'medium' ? '中等' : '困难'}</span>
                        <span class="task-reward">奖励: ${task.Reward} 绿宝石</span>
                    </div>
                    <div class="task-progress">
                        <span class="progress-text">已领取，等待完成</span>
                    </div>
                    <div class="task-actions">
                        <form action="/complete_task" method="post">
                            <input type="hidden" name="task_id" value="${task.ID}">
                            <button type="submit" class="minecraft-btn complete-btn">标记完成</button>
                        </form>
                    </div>
                </div>
            `;
        });
        claimedTasksContainer.innerHTML = html;
    }
    
    // 渲染即将开始任务
    const upcomingTasksContainer = document.getElementById('upcoming-tasks-container');
    if (upcomingTasksContainer) {
        let html = '';
        data.UpcomingTasks.forEach(task => {
            html += `
                <div class="task-card upcoming">
                    <h3 class="task-title">${task.Title}</h3>
                    <p class="task-description">${task.Description}</p>
                    <div class="task-meta">
                        <span class="task-difficulty difficulty-${task.Difficulty}">${task.Difficulty === 'easy' ? '简单' : task.Difficulty === 'medium' ? '中等' : '困难'}</span>
                        <span class="task-reward">奖励: ${task.Reward} 绿宝石</span>
                    </div>
                    <div class="task-start">开始时间: ${new Date(task.StartTime).toLocaleString()}</div>
                </div>
            `;
        });
        upcomingTasksContainer.innerHTML = html;
    }
    
    // 重新绑定事件
    bindTaskEvents();
}

// 加载商店数据
async function loadShopData() {
    try {
        const response = await fetch('/shop_data', {
            method: 'GET',
            credentials: 'include',
            headers: {
                'X-Requested-With': 'XMLHttpRequest'
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP错误 ${response.status}`);
        }
        
        const data = await response.json();
        
        if (data.success) {
            renderShop(data.data);
        } else {
            showMessage(data.message || '加载商店数据失败', 'error');
        }
    } catch (error) {
        console.error('加载商店数据失败:', error);
        showMessage('加载商店数据失败: ' + error.message, 'error');
    }
}

// 渲染商店数据
function renderShop(data) {
    // 更新玩家信息
    const playerNameElement = document.getElementById('player-name');
    const emeraldCountElement = document.getElementById('emerald-count');
    if (playerNameElement) playerNameElement.textContent = data.PlayerName;
    if (emeraldCountElement) emeraldCountElement.textContent = data.Emeralds;
    
    // 渲染物品列表
    const itemsContainer = document.getElementById('items-container');
    if (itemsContainer) {
        let html = '';
        data.Items.forEach(item => {
            html += `
                <div class="item-card">
                    <div class="item-image">
                        <img src="/static/images/item_${item.ID}.svg" alt="${item.Name}" onError="this.src='/static/images/default_item.svg'">
                    </div>
                    <h3 class="item-name">${item.Name}</h3>
                    <p class="item-description">${item.Description}</p>
                    <div class="item-meta">
                        <span class="item-cost">${item.Cost} 绿宝石</span>
                        <span class="item-stock">库存: ${item.Stock}</span>
                    </div>
                    <div class="item-actions">
                        <form action="/exchange" method="post">
                            <input type="hidden" name="item_id" value="${item.ID}">
                            <button type="submit" class="minecraft-btn exchange-btn" ${item.Stock <= 0 ? 'disabled' : ''}>
                                ${item.Stock <= 0 ? '库存不足' : '立即兑换'}
                            </button>
                        </form>
                    </div>
                </div>
            `;
        });
        itemsContainer.innerHTML = html;
    }
    
    // 重新绑定事件
    bindShopEvents();
}

// 加载管理员数据
async function loadAdminData() {
    try {
        const response = await fetch('/admin_data', {
            method: 'GET',
            credentials: 'include',
            headers: {
                'X-Requested-With': 'XMLHttpRequest'
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP错误 ${response.status}`);
        }
        
        const data = await response.json();
        
        if (data.success) {
            renderAdmin(data.data);
        } else {
            showMessage(data.message || '加载管理员数据失败', 'error');
        }
    } catch (error) {
        console.error('加载管理员数据失败:', error);
        showMessage('加载管理员数据失败: ' + error.message, 'error');
    }
}

// 渲染管理员数据
function renderAdmin(data) {
    // 渲染任务模板
    const taskTemplatesContainer = document.getElementById('task-templates-container');
    if (taskTemplatesContainer) {
        let html = '';
        data.TaskTemplates.forEach(template => {
            html += `
                <tr>
                    <td>${template.ID}</td>
                    <td>${template.Title}</td>
                    <td>${template.Difficulty === 'easy' ? '简单' : template.Difficulty === 'medium' ? '中等' : '困难'}</td>
                    <td>${template.Type === 'daily' ? '日常任务' : '限时任务'}</td>
                    <td>${template.Reward}</td>
                    <td>${template.RepeatDays}</td>
                    <td>
                        <form action="/delete_task_template" method="post" style="display: inline;">
                            <input type="hidden" name="template_id" value="${template.ID}">
                            <button type="submit" class="minecraft-btn small delete-btn">删除模板</button>
                        </form>
                    </td>
                </tr>
            `;
        });
        taskTemplatesContainer.innerHTML = html;
    }
    
    // 渲染任务实例
    const tasksContainer = document.getElementById('tasks-container');
    if (tasksContainer) {
        let html = '';
        data.Tasks.forEach(task => {
            html += `
                <tr>
                    <td>${task.ID}</td>
                    <td>${task.Title}</td>
                    <td>${task.Difficulty === 'easy' ? '简单' : task.Difficulty === 'medium' ? '中等' : '困难'}</td>
                    <td>${task.Type === 'daily' ? '日常任务' : '限时任务'}</td>
                    <td>${task.Reward}</td>
                    <td>
                        <span class="status-${task.Status}">
                            ${task.Status === 'available' ? '可领取' : task.Status === 'claimed' ? '已领取' : task.Status === 'completed' ? '已完成' : '已确认'}
                        </span>
                    </td>
                    <td>${task.ExpiryTime}</td>
                    <td>
                        ${task.Status === 'completed' ? `
                            <form action="/verify_task" method="post" style="display: inline;">
                                <input type="hidden" name="task_id" value="${task.ID}">
                                <button type="submit" class="minecraft-btn small">确认完成</button>
                            </form>
                        ` : ''}
                        <form action="/delete_task" method="post" style="display: inline;">
                            <input type="hidden" name="task_id" value="${task.ID}">
                            <button type="submit" class="minecraft-btn small delete-btn">删除</button>
                        </form>
                    </td>
                </tr>
            `;
        });
        tasksContainer.innerHTML = html;
    }
    
    // 渲染物品列表
    const itemsContainer = document.getElementById('items-container');
    if (itemsContainer) {
        let html = '';
        data.Items.forEach(item => {
            html += `
                <tr>
                    <td>${item.ID}</td>
                    <td>${item.Name}</td>
                    <td>${item.Description}</td>
                    <td>${item.Cost}</td>
                    <td>${item.Stock}</td>
                    <td><script>document.write(formatDateTime('${item.ExpiryTime}'))</script></td>
                    <td>
                        <form action="/update_item" method="post" style="display: inline;" id="update-item-form-${item.ID}">
                            <input type="hidden" name="item_id" value="${item.ID}">
                            <input type="hidden" name="name" value="${item.Name}" id="edit-name-${item.ID}">
                            <input type="hidden" name="description" value="${item.Description}" id="edit-description-${item.ID}">
                            <input type="hidden" name="cost" value="${item.Cost}" id="edit-cost-${item.ID}">
                            <input type="hidden" name="stock" value="${item.Stock}" id="edit-stock-${item.ID}">
                            <input type="hidden" name="expiry_time" value="${item.ExpiryTime}" id="edit-expiry-${item.ID}">
                            <button type="button" class="minecraft-btn small" onclick="window.openEditItemModal(${item.ID}, '${item.Name}', '${item.Description}', ${item.Cost}, ${item.Stock}, '${item.ExpiryTime}')">编辑</button>
                        </form>
                        <form action="/delete_item" method="post" style="display: inline;" id="delete-item-form-${item.ID}">
                            <input type="hidden" name="item_id" value="${item.ID}">
                            <button type="submit" class="minecraft-btn small danger">删除</button>
                        </form>
                    </td>
                </tr>
            `;
        });
        itemsContainer.innerHTML = html;
    }
    
    // 渲染兑换记录
    const exchangeRecordsContainer = document.getElementById('exchange-records-container');
    if (exchangeRecordsContainer) {
        let html = '';
        data.ExchangeRecords.forEach(record => {
            html += `
                <tr>
                    <td>${record.ID}</td>
                    <td>${record.PlayerID}</td>
                    <td>${record.ItemID}</td>
                    <td>${record.ItemName}</td>
                    <td>${record.Cost}</td>
                    <td><script>document.write(formatDateTime('${record.Timestamp}'))</script></td>
                    <td>
                        ${record.Exchanged ? `
                            <span class="status-verified">已兑换</span>
                        ` : `
                            <form action="/exchange_reward" method="post" style="display: inline;" id="exchange-form-${record.ID}">
                                <input type="hidden" name="exchange_id" value="${record.ID}">
                                <button type="submit" class="minecraft-btn small" id="exchange-btn-${record.ID}">兑换奖励</button>
                            </form>
                            <script>
                                // 使用立即执行函数表达式(IIFE)创建独立作用域，避免变量重复声明
                                (function() {
                                    // 为兑换按钮添加点击事件处理
                                    const exchangeBtn = document.getElementById('exchange-btn-${record.ID}');
                                    const exchangeForm = document.getElementById('exchange-form-${record.ID}');
                                    
                                    if (exchangeBtn && exchangeForm) {
                                        exchangeForm.addEventListener('submit', function(e) {
                                            // 禁用按钮并更改文本
                                            exchangeBtn.disabled = true;
                                            exchangeBtn.textContent = '已兑换';
                                            exchangeBtn.classList.add('disabled');
                                            // 不阻止表单提交，让请求继续处理
                                        });
                                    }
                                })();
                            </script>
                        `}
                    </td>
                </tr>
            `;
        });
        exchangeRecordsContainer.innerHTML = html;
    }
    
    // 重新绑定事件
    bindAdminEvents();
}

// 绑定任务相关事件
function bindTaskEvents() {
    // 为所有表单添加AJAX处理
    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
        // 通用成功回调
        const successCallback = function(data) {
            showMessage(data.message || '操作成功');
            // 刷新页面或更新内容
            if (data.refresh) {
                loadTasksData();
            }
        };
        
        // 通用错误回调
        const errorCallback = function(data) {
            showMessage(data.message || '操作失败', 'error');
        };
        
        // 为特定表单添加AJAX处理
        ajaxFormSubmit(form, successCallback, errorCallback);
    });
}

// 绑定商店相关事件
function bindShopEvents() {
    // 为所有表单添加AJAX处理
    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
        // 通用成功回调
        const successCallback = function(data) {
            showMessage(data.message || '操作成功');
            // 刷新页面或更新内容
            if (data.refresh) {
                loadShopData();
            }
        };
        
        // 通用错误回调
        const errorCallback = function(data) {
            showMessage(data.message || '操作失败', 'error');
        };
        
        // 为特定表单添加AJAX处理
        ajaxFormSubmit(form, successCallback, errorCallback);
    });
}

// 绑定管理员相关事件
function bindAdminEvents() {
    // 为所有表单添加AJAX处理
    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
        // 通用成功回调
        const successCallback = function(data) {
            showMessage(data.message || '操作成功');
            // 刷新页面或更新内容
            if (data.refresh) {
                loadAdminData();
            }
        };
        
        // 通用错误回调
        const errorCallback = function(data) {
            showMessage(data.message || '操作失败', 'error');
        };
        
        // 为特定表单添加AJAX处理
        ajaxFormSubmit(form, successCallback, errorCallback);
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
    
    // 根据当前页面加载对应的数据
    if (window.location.pathname === '/tasks') {
        // 加载任务数据
        loadTasksData();
    } else if (window.location.pathname === '/shop') {
        // 加载商店数据
        loadShopData();
    } else if (window.location.pathname === '/admin') {
        // 加载管理员数据
        loadAdminData();
    }

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