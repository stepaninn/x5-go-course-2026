package main

import (
	"errors"
	"fmt"

	"lecture5/practice"
)

func main() {
	fmt.Println("Сценарий 1")

	emailNotifier := &practice.EmailNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "email"},
		Address:      "test@example.com",
	}

	pushNotifier := &practice.PushNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "push"},
		DeviceID:     "test_id",
	}

	notifiers := []practice.Notifier{emailNotifier, pushNotifier}

	err := practice.Dispatch(notifiers, "Заказ готов")
	if err != nil {
		fmt.Println("Ожидалось успешная отправка: ", err)
	}

	fmt.Println("Email:", emailNotifier.Messages)
	fmt.Println("Push:", pushNotifier.Messages)

	fmt.Println("----------------")

	fmt.Println("Сценарий 2")

	err = practice.Dispatch(notifiers, "")

	if errors.Is(err, practice.ErrEmptyMessage) {
		fmt.Println("Пустое сообщение Dispatch")
	} else {
		fmt.Println("Ожидалась ошибка пустого сообщения")
	}

	fmt.Println("Email:", emailNotifier.Messages)
	fmt.Println("Push:", pushNotifier.Messages)

	fmt.Println("----------------")

	fmt.Println("Сценарий 3")

	err = practice.Dispatch(
		[]practice.Notifier{
			&practice.PushNotifier{
				BaseNotifier: practice.BaseNotifier{Channel: "push"},
			},
		},
		"Заказ готов",
	)

	var sendError *practice.SendError

	if errors.As(err, &sendError) {
		fmt.Printf("канал: %s\nпричина: %s\n", sendError.Channel, sendError.Reason)
	} else {
		fmt.Println("Ожидалась ошибка отправления сообщения")
	}

	fmt.Println("----------------")

	fmt.Println("Сценарий 4")
	errs := practice.SafeDispatch(
		[]practice.Notifier{
			&practice.BrokenNotifier{
				BaseNotifier: practice.BaseNotifier{Channel: "broken"},
			},
			pushNotifier,
		},
		"Сценарий 4",
	)

	for _, e := range errs {

		if errors.Is(e, practice.ErrNotifierPanic) {
			fmt.Printf("Ожидаемая паника Notifier: %v\n", e)
		} else {
			fmt.Println("Ожидалась паника")
		}
	}

	fmt.Println("Push: ", pushNotifier.Messages)

	fmt.Println("----------------")

	fmt.Println("Дополнительный сценарий")

	smsNotifier := &practice.SMSNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "Extra"},
		Number:       "+7 999 999 99 99",
	}

	err = practice.Dispatch(
		[]practice.Notifier{
			smsNotifier,
		},
		"Доп",
	)

	if err != nil {
		fmt.Println("Неожиданная ошибка", err)
	}

	fmt.Println("Extra:", smsNotifier.Messages)
}
