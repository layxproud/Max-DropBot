package messages

const (
	Start = "Вас приветствует бот для загрузки файлов на сервер SAMPLE_NAME." +
		" Бот принимает файлы до 25 МБ следующих форматов:\n" +
		"1) PDF (.pdf)\n2) PowerPoint (.ppt, .pptx)\n" +
		"3) Microsoft Word: (.doc, .docx)\n" +
		"4) Видео форматы (.mp4, .avi, .mov, .mkv, .wmv)\n" +
		"5) Аудио форматы (.mp3, .wav, .aac, .m4a, .flac)\n" +
		"6) Изображения (.png, .jpg, .jpeg, .gif, .tiff, .svg)\n\n" +
		"Для продолжения просто пришлите файл в этот чат."

	Info = "Бот принимает файлы:\n" +
		"1) PDF (.pdf)\n2) PowerPoint (.ppt, .pptx)\n" +
		"3) Microsoft Word: (.doc, .docx)\n" +
		"4) Видео форматы (.mp4, .avi, .mov, .mkv, .wmv)\n" +
		"5) Аудио форматы (.mp3, .wav, .aac, .m4a, .flac)\n" +
		"6) Изображения (.png, .jpg, .jpeg, .gif, .tiff, .svg)"

	UnknownCommand = "Неизвестная команда. Список команд:\n1) /info\n2) /start"

	FileUploaded = "✅ Файл %s загружен!\n%s"

	FileTooLarge = "❌ Файл %s превышает ограничение в 25 МБ!"

	UnsupportedFormat = "❌ Расширение файла %s не поддерживается!\n" +
		"Для получения справки по принимаемым форматам напишите /info"

	InternalError = "❌ Не удалось скачать файл %s из-за внутренней ошибки"
)
