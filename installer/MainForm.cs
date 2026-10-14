using System;
using System.Drawing;
using System.IO;
using System.Runtime.InteropServices;
using System.Windows.Forms;

namespace LuminaInstaller
{
    internal sealed class MainForm : Form
    {
        private const string AppVersion = Installer.AppVersion;

        private static readonly Color Bg = Color.FromArgb(30, 30, 30);
        private static readonly Color Panel = Color.FromArgb(43, 43, 46);
        private static readonly Color Fg = Color.FromArgb(230, 230, 230);
        private static readonly Color Sub = Color.FromArgb(150, 150, 150);
        private static readonly Color Accent = Color.FromArgb(14, 99, 156);
        private static readonly Color Border = Color.FromArgb(63, 63, 70);

        private readonly TextBox _pathBox;
        private readonly CheckBox _pathCheck;
        private readonly Label _status;
        private readonly Button _installBtn;
        private readonly Button _cancelBtn;
        private readonly ProgressBar _bar;

        public MainForm()
        {
            Text = "LuminaLang 安装程序";
            ClientSize = new Size(560, 340);
            BackColor = Bg;
            ForeColor = Fg;
            Font = new Font("Microsoft YaHei UI", 9f);
            FormBorderStyle = FormBorderStyle.FixedDialog;
            MaximizeBox = false;
            MinimizeBox = false;
            StartPosition = FormStartPosition.CenterScreen;

            var title = new Label
            {
                Text = "LuminaLang",
                Font = new Font("Segoe UI", 20f, FontStyle.Bold),
                ForeColor = Fg,
                AutoSize = true,
                Location = new Point(26, 20),
            };

            var sub = new Label
            {
                Text = "编译器安装程序  ·  v" + AppVersion,
                ForeColor = Sub,
                AutoSize = true,
                Location = new Point(30, 62),
            };

            var locLabel = new Label
            {
                Text = "安装位置",
                ForeColor = Fg,
                AutoSize = true,
                Location = new Point(30, 106),
            };

            _pathBox = new TextBox
            {
                Location = new Point(30, 130),
                Width = 396,
                BackColor = Panel,
                ForeColor = Fg,
                BorderStyle = BorderStyle.FixedSingle,
                Text = Installer.DefaultDir(),
            };

            var browse = MakeButton("浏览...", 436, 129, 96, 26);
            browse.Click += (_, _) => Browse();

            _pathCheck = new CheckBox
            {
                Text = "将安装目录添加到 PATH（当前用户）",
                ForeColor = Fg,
                AutoSize = true,
                Location = new Point(30, 174),
                Checked = true,
            };

            _bar = new ProgressBar
            {
                Location = new Point(30, 214),
                Size = new Size(502, 6),
                Style = ProgressBarStyle.Marquee,
                MarqueeAnimationSpeed = 30,
                Visible = false,
            };

            _status = new Label
            {
                Text = "选择安装位置后点击“安装”。",
                ForeColor = Sub,
                AutoSize = true,
                Location = new Point(30, 228),
            };

            _cancelBtn = MakeButton("取消", 436, 280, 96, 32);
            _cancelBtn.Click += (_, _) => Close();

            _installBtn = MakeButton("安装", 332, 280, 96, 32);
            _installBtn.BackColor = Accent;
            _installBtn.ForeColor = Color.White;
            _installBtn.FlatAppearance.BorderSize = 0;
            _installBtn.Click += (_, _) => Install();

            AcceptButton = _installBtn;
            CancelButton = _cancelBtn;

            Controls.AddRange(new Control[]
            {
                title, sub, locLabel, _pathBox, browse, _pathCheck, _bar, _status, _cancelBtn, _installBtn,
            });
        }

        private Button MakeButton(string text, int x, int y, int w, int h)
        {
            var b = new Button
            {
                Text = text,
                Location = new Point(x, y),
                Size = new Size(w, h),
                FlatStyle = FlatStyle.Flat,
                BackColor = Panel,
                ForeColor = Fg,
                Font = new Font("Microsoft YaHei UI", 9f),
            };
            b.FlatAppearance.BorderColor = Border;
            b.FlatAppearance.MouseOverBackColor = Color.FromArgb(60, 60, 64);
            return b;
        }

        private void Browse()
        {
            using var dialog = new FolderBrowserDialog
            {
                Description = "选择 LuminaLang 的安装位置",
                UseDescriptionForTitle = true,
                ShowNewFolderButton = true,
                SelectedPath = Directory.Exists(_pathBox.Text) ? _pathBox.Text : Installer.DefaultDir(),
            };
            if (dialog.ShowDialog(this) == DialogResult.OK)
            {
                _pathBox.Text = dialog.SelectedPath;
            }
        }

        private void Install()
        {
            string dir = _pathBox.Text.Trim().Trim('"');
            if (dir.Length == 0)
            {
                MessageBox.Show(this, "请先选择安装位置。", "LuminaLang", MessageBoxButtons.OK, MessageBoxIcon.Warning);
                return;
            }

            _installBtn.Enabled = false;
            _cancelBtn.Enabled = false;
            _bar.Visible = true;
            Application.DoEvents();

            try
            {
                Installer.Install(dir, _pathCheck.Checked, SetStatus);
                _bar.Visible = false;

                string message = "LuminaLang 已安装到：\n" + dir + "\n\n"
                    + "请确保已安装 LLVM/clang 并加入 PATH（或用 LUMINA_CC 指定 clang 路径）。";
                if (_pathCheck.Checked)
                {
                    message += "\n\nPATH 已更新，新开的终端即可使用 luminalang 与 pho 命令。";
                }

                MessageBox.Show(this, message, "安装完成", MessageBoxButtons.OK, MessageBoxIcon.Information);
                Close();
            }
            catch (Exception ex)
            {
                _bar.Visible = false;
                SetStatus("安装失败。");
                MessageBox.Show(this, "安装失败：\n" + ex.Message, "LuminaLang", MessageBoxButtons.OK, MessageBoxIcon.Error);
                _installBtn.Enabled = true;
                _cancelBtn.Enabled = true;
            }
        }

        private void SetStatus(string text)
        {
            _status.Text = text;
            _status.Refresh();
            Application.DoEvents();
        }

        [DllImport("dwmapi.dll")]
        private static extern int DwmSetWindowAttribute(IntPtr hwnd, int attribute, ref int value, int size);

        protected override void OnHandleCreated(EventArgs e)
        {
            base.OnHandleCreated(e);
            int enabled = 1;
            DwmSetWindowAttribute(Handle, 20, ref enabled, sizeof(int));
        }
    }
}
